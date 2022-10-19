package sinoname

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

	"golang.org/x/sync/errgroup"
)

// transformerLayer holds all the transformers belonging to it (statefull or not),
// when the layer runs it fans out all the messages it gets to all
// the transformers it owns (first to the unstatefull then to the statefull).
//
// teoretically 1 message to a layer with 4 transformers results in 4 messages (1 * 4).
type TransformerLayer struct {
	cfg                  *Config
	init                 int32
	transformers         []Transformer
	transformerFactories []TransformerFactory
}

// PumpOut recieves messages from the upstream layer via the in channel and passes them through the transformers.
// The end products of the transformers are fed in the returned channel.
func (l *TransformerLayer) PumpOut(ctx context.Context, g *errgroup.Group, in <-chan string) (<-chan string, error) {
	if len(l.transformers) == 0 && len(l.transformerFactories) == 0 {
		return nil, errors.New("sinoname: layer has no transformers")
	}

	// local copy of statefull trasnformers.
	statefullTransformers := l.getStatefullTransformers()

	outC := make(chan string)
	// wg is used to monitor the local go routines of this layer.
	var wg sync.WaitGroup
	pumpOut := func(ctx context.Context, t Transformer, v string) func() error {
		f := func() error {
			defer wg.Done()

			val, err := t.Transform(ctx, v)
			if err != nil {
				if err == ErrSkip {
					return nil
				}
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
		// upstream's out channel closure, cleanup.
		defer func() {
			// wait for all the transformers to send their value before closing.
			wg.Wait()
			close(outC)
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

				wg.Add(len(l.transformers) + len(statefullTransformers))
				for _, t := range l.transformers {
					g.Go(
						pumpOut(ctx, t, v),
					)
				}

				for _, t := range statefullTransformers {
					g.Go(
						pumpOut(ctx, t, v),
					)
				}
			}
		}
	}()

	return outC, nil
}

func (l *TransformerLayer) getStatefullTransformers() []Transformer {
	if len(l.transformerFactories) == 0 {
		return nil
	}

	statefullTransformers := make([]Transformer, len(l.transformerFactories))
	// get initiall values if the first caller.
	if atomic.CompareAndSwapInt32(&l.init, 0, 1) {
		diff := len(l.transformers) - len(l.transformerFactories)
		copy(statefullTransformers, l.transformers[diff:])
		l.transformers = l.transformers[:diff]
		return statefullTransformers
	}

	for i, f := range l.transformerFactories {
		t, _ := f(l.cfg)
		statefullTransformers[i] = t
	}
	return statefullTransformers
}
