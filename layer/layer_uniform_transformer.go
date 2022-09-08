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

	outC := make(chan string, len(l.Transformers))
	buf := newSyncBuf(len(l.Transformers), outC)
	broadcast := &messageBroadcast{
		source:    in,
		g:         g,
		listeners: make([]chan string, 0),
	}

	pumpToBuf := func(in <-chan string) {
		for {
			val, ok := <-in
			if !ok {
				return
			}

			// blocks till all the other pumps have wrote their message.
			buf.write(val)
		}
	}

	// register transformers for broadcast.
	for _, t := range l.Transformers {
		tOut := broadcast.register(t)
		go pumpToBuf(tOut)
	}

	go broadcast.start(ctx)

	return outC, nil
}

type syncBuffer struct {
	nWriters int
	isClosed bool
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

func (b *syncBuffer) close() {
	b.c.L.Lock()
	defer b.c.L.Unlock()

	b.isClosed = true
	b.c.Broadcast()
}

func (b *syncBuffer) write(val string) {
	b.c.L.Lock()
	defer b.c.L.Unlock()

	if b.isClosed {
		return
	}

	b.buf = append(b.buf, val)
	// last value written.
	if len(b.buf) == b.nWriters {
		// share the values.
		b.flush()
		// wake up other writers.
		b.c.Broadcast()
		return
	}

	b.c.Wait()
}

func (b *syncBuffer) flush() {
	for _, v := range b.buf {
		b.ch <- v
	}
	b.buf = b.buf[:0]
}

// messageBroadcast takes responsability of layer source and broadcasts it to all the
// transformers returning a channel which is buffered to read the results from so that
// the layer doesent waste time, it instead prepares values for the next synced write.
type messageBroadcast struct {
	source    <-chan string
	wg        sync.WaitGroup
	g         *errgroup.Group
	listeners []chan string
}

// register must be called before starting the service.
func (m *messageBroadcast) register(t transformer.Transformer) <-chan string {
	outC := make(chan string, 10) // buffer to prevent blocking.
	inC := make(chan string)

	m.g.Go(m.handleTransformer(inC, outC, t))

	return outC
}

func (m *messageBroadcast) start(ctx context.Context) {
	// defer func() {

	// }

	for {
		select {
		case <-ctx.Done():
			return

		case val, ok := <-m.source:
			if !ok {
				return
			}

			for _, listener := range m.listeners {
				select {
				case listener <- val:
				case <-ctx.Done():
					return
				}
			}
		}
	}
}

func (m *messageBroadcast) handleTransformer(in <-chan string, out chan<- string, t transformer.Transformer) func() error {
	f := func() error {

	}

	return f
}
