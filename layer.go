package sinoname

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"golang.org/x/sync/errgroup"
)

// Layers is an abstraction type for multiple layers.
type Layers []*Layer

// Run runs all the layers it owns returning a channel to read from and a cleanup function
// which must be called in order to free all resources.
func (s Layers) Run(ctx context.Context, in string) (<-chan string, func() error, error) {
	if len(s) == 0 {
		return nil, func() error { return nil }, fmt.Errorf("sinoname: generator has no layers")
	}

	// fanOutC is used to fanOut messages to the first layer.
	fanOutC := make(chan string, 1)
	fanOutC <- in
	close(fanOutC)

	// the errgroup is used to stop all layers once one of the layers
	// encounters an error from the source, rendering the source unreliable.
	g, ctx := errgroup.WithContext(ctx)

	// cancel is used to cancel the layers when the generator is done reading messages from the
	// pipeline freeing the go-routines.
	ctx, cancel := context.WithCancel(ctx)

	// fanInC is used to fanIn all the messages from the last layer for the generator to consume.
	//
	// it is closed when either the context is cancelled by an error or explicit cancelation by the client
	// or generator.
	var fanInC <-chan string

	// clnUp frees all go-routines created by the layers, not calling this function can cause
	// a memory leak.
	clnUp := func() error {
		// cancel the child context, may already be cancelled (error ocurred or layers finished)
		cancel()
		// wait for layer go-routines to exit.
		err := g.Wait()
		if err == context.Canceled {
			return nil
		}
		return err
	}

	// start layers pipeline.
	var err error
	var lastOutC <-chan string
	lastOutC = fanOutC
	for _, layer := range s {
		lastOutC, err = layer.Run(ctx, g, lastOutC)
		if err != nil {
			return nil, clnUp, err
		}
	}
	fanInC = lastOutC

	return fanInC, clnUp, nil
}

// Layer holds all the transformers belonging to it, when the layer runs it fans out all the
// messages it gets to all the transformers it owns.
//
// teoretically 1 message to a layer with 4 transformers results in 4 messages. 1 * 4
// the output formula of each layer is sum(messages from up stream layer) * len(transformers)
type Layer struct {
	// cfg points to the config it belongs to.
	// used to access meta data.
	cfg *Config

	// transformers is the transformers which get run for each message from the upstream channel.
	transformers []Transformer
}

// Run recieves messages from the upstream layer via the in channel and passes them through the transformers.
// The end products of the transformers are fed in the returned channel.
func (s *Layer) Run(ctx context.Context, g *errgroup.Group, in <-chan string) (<-chan string, error) {
	outC := make(chan string, s.cfg.MaxVals)

	if len(s.transformers) == 0 {
		return nil, errors.New("sinoname: layer has no transformers")
	}

	// wg is used to monitor the local go routines of this layer.
	var wg sync.WaitGroup
	pumpOut := func(ctx context.Context, t Transformer, v string) func() error {
		f := func() error {
			defer wg.Done()
			val, err := t.Transform(v)
			if err != nil {
				return err
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case outC <- val:
				return nil
			}
		}

		return f
	}

	go func() {
		// before the factory go-routine exits, either by a context cancelation or by the
		// upstream's out channel closure, close the layers out channel.
		defer func() {
			defer close(outC)
			// wait for all the transformers to send their value before closing.
			wg.Wait()
		}()

		for {
			select {
			// ctx close signals a deliberate close or that an error occured somewhere
			// throughout the pipeline, eitherway stop layer.
			case <-ctx.Done():
				return
			case v, ok := <-in:
				if !ok {
					return
				}

				wg.Add(len(s.transformers))
				for _, t := range s.transformers {
					g.Go(
						pumpOut(ctx, t, v),
					)
				}
			}
		}
	}()

	return outC, nil
}

// LayerFactory creates a new layer with the given config.
type LayerFactory func(cfg *Config) *Layer

// Pack merges n layers into one singel layer, when a message is recieved by the layer it gets
// distributed to all the transformers it has.
//
// WARNING: Pack panics if it gets called with a Layer which has multiple transformers, the
// max number of transformers per layer passed to pack is 1.
func Pack(layerFacts ...LayerFactory) LayerFactory {
	return func(cfg *Config) *Layer {
		layer := &Layer{
			cfg:          cfg,
			transformers: make([]Transformer, len(layerFacts)),
		}

		// populate new layer with transformers of other layers.
		for i, layerFact := range layerFacts {
			l := layerFact(cfg)
			if len(l.transformers) != 1 {
				panic("sinoname: unexpected number of transformers in layer")
			}
			layer.transformers[i] = l.transformers[0]
		}

		return layer
	}
}
