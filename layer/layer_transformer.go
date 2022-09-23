package layer

import (
	"context"
	"errors"
	"sync"

	"github.com/Lambels/sinoname/transformer"
	"golang.org/x/sync/errgroup"
)

// transformerLayer holds all the transformers belonging to it, when the layer runs it fans out all the
// messages it gets to all the transformers it owns.
//
// teoretically 1 message to a layer with 4 transformers results in 4 messages (1 * 4).
// the output formula of each layer is sum(messages from up stream layer) * len(transformers)
type TransformerLayer struct {
	// transformers is the transformers which get run for each message from the upstream channel.
	Transformers []transformer.Transformer
}

// PumpOut recieves messages from the upstream layer via the in channel and passes them through the transformers.
// The end products of the transformers are fed in the returned channel.
func (l *TransformerLayer) PumpOut(ctx context.Context, g *errgroup.Group, in <-chan string) (<-chan string, error) {
	if len(l.Transformers) == 0 {
		return nil, errors.New("sinoname: layer has no transformers")
	}

	outC := make(chan string)
	// wg is used to monitor the local go routines of this layer.
	var wg sync.WaitGroup
	pumpOut := func(ctx context.Context, t transformer.Transformer, v string) func() error {
		f := func() error {
			defer wg.Done()

			val, err := t.Transform(ctx, v)
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
			if ctx.Err() != nil {
				return
			}

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

				wg.Add(len(l.Transformers))
				for _, t := range l.Transformers {
					g.Go(
						pumpOut(ctx, t, v),
					)
				}
			}
		}
	}()

	return outC, nil
}
