package sinoname

import (
	"context"
	"errors"

	"golang.org/x/sync/errgroup"
)

// Layer represents an abstraction for a layer in the processing pipeline.
//
// Producer -> Layer 1 -> Layer 2 -> Layer 3 -> Sink
//
// Each layer in the pipeline must implement the PumpOut method which is responsible for
// consuming values from a channel (which will close and must be handeled), modifing the
// value and then passing the value back out the returned channel.
//
// Producer -> Layer 1 -> Layer 2 -> Layer 3 -> Sink
//
// Here layer 1 consumes 1 value from the producer which automatically closes its channel after
// the one value it sends, then layer 1 is responsible for handling the value it gets, sending
// the value out (optional) and closing its out channel.
//
// A layer must close its out channel if:
//
// 1: The context is cancelled. This may be caused by the client or by and error through the
// pipeline.
//
// 2: The upstream channel is closed, this means that the upstream layer signaled that its done
// sending values, process all the messages, send them out and close the out channel.
//
// All processing inside the layer must be called in the *errgroup.Group passed in via:
//
//	// to close the pipeline return an error here.
//	(*errgroup.Group).Go(func() error {})
type Layer interface {
	// PumpOut handles reads messages from the channel, processes them and sends them through
	// the out channel.
	PumpOut(context.Context, *errgroup.Group, <-chan string) (<-chan string, error)
}

// LayerFactory takes in a config object and returns a new layer.
type LayerFactory func(cfg *Config) Layer

// Layers is an abstraction type for multiple layers.
type Layers []Layer

// Run runs all the layers it owns returning a channel to read from and a cleanup function
// which must be called in order to free all resources.
func (s Layers) Run(ctx context.Context, in string) (<-chan string, func() error, error) {
	if len(s) == 0 {
		return nil, func() error { return nil }, errors.New("sinoname: generator has no layers")
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

	// clnUp frees all go-routines created by the layers, not calling this function can cause
	// a memory leak.
	clnUp := func() error {
		// cancel the child context, may already be cancelled (error ocurred)
		cancel()
		// wait for layer go-routines to exit.
		return g.Wait()
	}

	// fanInC is used to fanIn all the messages from the last layer for the generator to consume.
	//
	// it is closed when either the context is cancelled by an error or explicit cancelation by the client
	// or generator.
	var fanInC <-chan string
	// start layers pipeline.
	var err error
	var lastOutC <-chan string
	lastOutC = fanOutC
	for _, layer := range s {
		lastOutC, err = layer.PumpOut(ctx, g, lastOutC)
		if err != nil {
			return nil, clnUp, err
		}
	}
	fanInC = lastOutC

	return fanInC, clnUp, nil
}
