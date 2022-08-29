package sinoname

import (
	"context"
	"errors"
	"sync"

	"golang.org/x/sync/errgroup"
)

// transformerLayer holds all the transformers belonging to it, when the layer runs it fans out all the
// messages it gets to all the transformers it owns.
//
// teoretically 1 message to a layer with 4 transformers results in 4 messages. 1 * 4
// the output formula of each layer is sum(messages from up stream layer) * len(transformers)
type transformerLayer struct {
	// cfg points to the config it belongs to.
	// used to access meta data.
	cfg *Config

	// transformers is the transformers which get run for each message from the upstream channel.
	transformers []Transformer
}

// PumpOut recieves messages from the upstream layer via the in channel and passes them through the transformers.
// The end products of the transformers are fed in the returned channel.
func (l *transformerLayer) PumpOut(ctx context.Context, g *errgroup.Group, in <-chan string) (<-chan string, error) {
	outC := make(chan string)

	if len(l.transformers) == 0 {
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

				wg.Add(len(l.transformers))
				for _, t := range l.transformers {
					g.Go(
						pumpOut(ctx, t, v),
					)
				}
			}
		}
	}()

	return outC, nil
}
