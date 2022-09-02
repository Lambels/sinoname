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
	buf := newSyncBuf(len(l.Transformers))

	pumpOutOnSync := func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			}

			vals := buf.sync()
			for _, v := range vals {
				outC <- v
			}
		}
	}

	go func() {

		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-in:
				if !ok {
					return
				}

				// generate the value in local buffer.
			}
		}
	}()
}

type token struct{}

type syncFlushBuffer struct {
	sem chan token
	wg  sync.WaitGroup
	buf []string
}

func newSyncBuf(maxBuf int) *syncFlushBuffer {
	b := &syncFlushBuffer{
		sem: make(chan token, maxBuf),
		buf: make([]string, 0),
	}
	return b
}

func (b *syncFlushBuffer) send(val string) {
	b.sem <- token{}

	b.buf = append(b.buf, val)
	if len(b.sem) == cap(b.sem) {
		b.wg.Done()
		return
	}

	b.wg.Wait()
}

func (b *syncFlushBuffer) sync() []string {
	b.wg.Wait()

	bufCopy := make([]string, len(b.buf))
	copy(bufCopy, b.buf)
	b.buf = b.buf[:0]

	b.wg.Add(1)
	for len(b.sem) > 0 {
		<-b.sem
	}
	return bufCopy
}
