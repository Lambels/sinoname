package layer

import (
	"context"
	"errors"
	"sync"

	"github.com/Lambels/sinoname/transformer"
	"golang.org/x/sync/errgroup"
)

type UniformTransformerLayer struct {
	Transformers []transformer.Transformer
}

func (l *UniformTransformerLayer) PumpOut(ctx context.Context, g *errgroup.Group, in <-chan string) (<-chan string, error) {
	if len(l.Transformers) == 0 {
		return nil, errors.New("sinoname: layer has no transformers")
	}

	outC := make(chan string)
	buf := newSyncBuf(len(l.Transformers), outC)

	// wg is used to monitor the local go routines of this layer.
	var wg sync.WaitGroup
	// pumpToBuf pumps messages to the sync buffer to be dispatched in batches.
	pumpToBuf := func(ctx context.Context, t transformer.Transformer, v string) func() error {
		f := func() error {
			defer wg.Done()
			val, err := t.Transform(v)
			if err != nil {
				return err
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-buf.writeWithSignal(val):
				return nil
			}
		}

		return f
	}

	go func() {
		defer func() {
			defer close(outC)

			wg.Wait()
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-in:
				if !ok {
					return
				}

				wg.Add(len(l.Transformers))
				for _, t := range l.Transformers {
					g.Go(
						pumpToBuf(ctx, t, v),
					)
				}
			}
		}
	}()

	return outC, nil
}

type syncBuffer struct {
	nWriters int
	c        *sync.Cond
	ch       chan<- string
	buf      []string
}

func newSyncBuf(n int, out chan<- string) *syncBuffer {
	b := &syncBuffer{
		nWriters: n,
		c:        sync.NewCond(&sync.Mutex{}),
		ch:       out,
		buf:      make([]string, 0),
	}
	return b
}

func (b *syncBuffer) writeWithSignal(val string) <-chan struct{} {
	ch := make(chan struct{})

	go func() {
		b.write(val)
		ch <- struct{}{}
	}()

	return ch
}

func (b *syncBuffer) write(val string) {
	b.c.L.Lock()
	defer b.c.L.Unlock()

	b.buf = append(b.buf, val)
	// last value written
	if len(b.buf) == b.nWriters {
		// share the values.
		b.share()
		// wake up other writers.
		b.c.Broadcast()
		return
	}

	b.c.Wait()
}

// share must be called with an aquired lock.
func (b *syncBuffer) share() {
	for _, v := range b.buf {
		b.ch <- v
	}
	b.buf = b.buf[:0]
}
