package layer

import (
	"context"
	"errors"
	"sync"

	"github.com/Lambels/sinoname/transformer"
	"golang.org/x/sync/errgroup"
)

// UniformTransformerLayer syncronises writes to the the downsream layer whilst not blocking
// the upstream layer.
//
// UniformTransformerLayer should be used when having transformers with very different
// transforming speeds. Instead of flooding the pipeline with lower speed messages from
// faster transformers the UniformTransformerLayer waits for the slower and faster transformers
// to write at the same time.
//
// Even though the layer waits for the messages to be written at the same time it doesent
// "sleep". Each transformer has a buffer so that when new messages come and a previous
// message is synced, the layer doesent wait for the previous message to be synced, it takes
// the new message, processes it and then writes it to the buffer. When the previous message
// is synced the layer pulls messages from each transformers buffer and syncs them.
type UniformTransformerLayer struct {
	Transformers []transformer.Transformer
}

func (l *UniformTransformerLayer) PumpOut(ctx context.Context, g *errgroup.Group, in <-chan string) (<-chan string, error) {
	if len(l.Transformers) == 0 {
		return nil, errors.New("sinoname: layer has no transformers")
	}

	outC := make(chan string, len(l.Transformers))
	buf := newSyncBuf(len(l.Transformers), outC)
	// this wg is shared by the message broadcaster and message consumer,
	// the broadcaster increments it whilst the consumer decrements it.
	//
	// this allows the broadcaster to wait for messages to be written to the sync buffer
	// before closing it.
	wg := &sync.WaitGroup{}
	broadcast := newMessageBroadcast(ctx, in, g, buf, wg)

	// go routine which reads from pumpIn channel buffer and writes to the sync buffer,
	// once the value is written the wg is decremented.
	pumpToBuf := func(pumpIn <-chan string, wg *sync.WaitGroup) {
		for val := range pumpIn {
			// blocks till all the other pumps have wrote their message.
			// once the buffer is closed, the write value wont block.
			buf.write(val)
			wg.Done()
		}
	}

	// register transformers for broadcast.
	for _, t := range l.Transformers {
		tOut := broadcast.register(t)
		go pumpToBuf(tOut, wg)
	}

	go broadcast.start()

	return outC, nil
}

// syncBuffer waits for nWriters to write their value via the (*syncBuffer).write(),
// each writer should call the write function in its own go-routine.
type syncBuffer struct {
	nWriters int
	isClosed bool
	// monitor isClosed and buf values.
	c   *sync.Cond
	ch  chan<- string
	buf []string
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

// close closes the out channel and makes write calls stop blocking and no-op.
func (b *syncBuffer) close() {
	b.c.L.Lock()
	defer b.c.L.Unlock()

	b.isClosed = true
	close(b.ch)
	b.c.Broadcast()
}

// write writes one value to the buf and then waits for the other write calls to write their
// value then unblocks.
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

// must be called with an aquired lock.
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
	buf       *syncBuffer
	wg        *sync.WaitGroup
	g         *errgroup.Group
	ctx       context.Context
	listeners []chan<- string
}

func newMessageBroadcast(ctx context.Context, source <-chan string, g *errgroup.Group, buf *syncBuffer, wg *sync.WaitGroup) *messageBroadcast {
	return &messageBroadcast{
		source:    source,
		buf:       buf,
		g:         g,
		wg:        wg,
		ctx:       ctx,
		listeners: make([]chan<- string, 0),
	}

}

// register must be called before starting the service.
func (m *messageBroadcast) register(t transformer.Transformer) <-chan string {
	outC := make(chan string, 10) // buffer to prevent blocking.
	inC := make(chan string)
	m.listeners = append(m.listeners, inC)

	go m.handleTransformer(inC, outC, t)

	return outC
}

func (m *messageBroadcast) start() {
	defer func() {
		defer m.buf.close()
		// close all listeners -> close all transformer handlers.
		// the handlers will wait in their go routines for the buffer to accept values.
		m.close()

		if m.ctx.Err() != nil {
			return
		}

		m.wg.Wait()
	}()

	for {
		select {
		case <-m.ctx.Done():
			return

		case val, ok := <-m.source:
			if !ok {
				return
			}

			for _, listener := range m.listeners {
				select {
				case listener <- val:
					m.wg.Add(1)
				case <-m.ctx.Done():
					return
				}
			}
		}
	}
}

// handleTransformer pumps messages to the transformers out channel concurrently.
func (m *messageBroadcast) handleTransformer(in <-chan string, out chan<- string, t transformer.Transformer) {
	defer func() {
		m.wg.Wait()
		close(out)
	}()

	for val := range in {
		m.g.Go(m.pumpToOut(val, t, out))
	}
}

func (m *messageBroadcast) pumpToOut(val string, t transformer.Transformer, out chan<- string) func() error {
	f := func() error {
		select {
		case <-m.ctx.Done():
			return nil
		case sig := <-transformer.TransformWithSignal(t, val):
			if sig.Err != nil {
				return sig.Err
			}
			out <- sig.Val
		}

		return nil
	}

	return f
}

func (m *messageBroadcast) close() {
	for _, listener := range m.listeners {
		close(listener)
	}
}
