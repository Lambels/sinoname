package sinoname

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

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
	cfg                  *Config
	init                 int32
	transformers         []Transformer
	transformerFactories []TransformerFactory
}

func (l *UniformTransformerLayer) PumpOut(ctx context.Context, g *errgroup.Group, in <-chan string) (<-chan string, error) {
	if len(l.transformers) == 0 && len(l.transformerFactories) == 0 {
		return nil, errors.New("sinoname: layer has no transformers")
	}

	// local copy of statefull trasnformers.
	statefullTransformers := l.getStatefullTransformers()

	outC := make(chan string)
	out := newSyncOut(len(l.transformers)+len(l.transformerFactories), outC)
	// this wg is shared by the message broadcaster and message consumer,
	// it acts as the orchestrator between the two,
	// the broadcaster increments it whilst the consumer decrements it.
	//
	// this allows the broadcaster to wait for messages to be written to the sync buffer
	// before closing it.
	wg := &sync.WaitGroup{}
	broadcast := newMessageBroadcast(ctx, in, g, out, wg)

	// go routine which reads from pumpIn channel buffer and writes to the sync buffer,
	// once the value is written the wg is decremented.
	pumpToSyncBuf := func(pumpIn <-chan string, skip <-chan struct{}, id int, wg *sync.WaitGroup) {
		for {
			select {
			case <-ctx.Done():
				return

			case val, ok := <-pumpIn:
				if !ok {
					return
				}

				next := out.write(id, val)
				wg.Done()
				if !next {
					return
				}

			case <-skip:
				next := out.advance(id)
				wg.Done()
				if !next {
					return
				}
			}
		}
	}

	// register transformers for broadcast.
	for i, t := range l.transformers {
		tOut, tSkip := broadcast.register(t)
		go pumpToSyncBuf(tOut, tSkip, i, wg)
	}

	// register statefull transformers for broadcast.
	for i, t := range statefullTransformers {
		id := len(l.transformers) + i - 1
		tOut, tSkip := broadcast.register(t)
		go pumpToSyncBuf(tOut, tSkip, id, wg)
	}

	go broadcast.start()

	return outC, nil
}

// messageBroadcast takes responsability of layer source and broadcasts it to all the
// transformers returning a channel which is buffered to read the results from so that
// the layer doesent waste time, it instead prepares values for the next synced write.
type messageBroadcast struct {
	source    <-chan string
	out       *syncOut
	wg        *sync.WaitGroup
	g         *errgroup.Group
	ctx       context.Context
	listeners []chan<- string
}

func newMessageBroadcast(ctx context.Context, source <-chan string, g *errgroup.Group, out *syncOut, wg *sync.WaitGroup) *messageBroadcast {
	return &messageBroadcast{
		source:    source,
		out:       out,
		g:         g,
		wg:        wg,
		ctx:       ctx,
		listeners: make([]chan<- string, 0),
	}

}

func (m *messageBroadcast) start() {
	// before the factory go-routine exits, either by a context cancelation or by the
	// upstream's out channel closure, cleanup.
	defer func() {
		// close all listeners -> close all transformer handlers.
		// the handlers will wait in their go routines for the buffer to accept values.
		m.close()

		// if ctx cancelled just close buffer imediately, this will clean up any go-routine.
		if m.ctx.Err() != nil {
			m.out.close()
			return
		}

		m.wg.Wait()
		m.out.close()
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
				case <-m.ctx.Done():
					return
				default:
				}

				listener <- val
				m.wg.Add(1)
			}
		}
	}
}

// register must be called before starting the service.
func (m *messageBroadcast) register(t Transformer) (<-chan string, <-chan struct{}) {
	outC := make(chan string, 10) // buffer to prevent blocking.
	skip := make(chan struct{}, 10)
	inC := make(chan string)
	m.listeners = append(m.listeners, inC)

	go m.handleTransformer(inC, outC, skip, t)

	return outC, skip
}

// handleTransformer pumps messages to the transformers out channel concurrently.
func (m *messageBroadcast) handleTransformer(in <-chan string, out chan<- string, skip chan<- struct{}, t Transformer) {
	defer func() {
		m.wg.Wait()
		close(out)
	}()

	for val := range in {
		m.g.Go(m.pumpToOut(val, t, out, skip))
	}
}

// pumpToOut carries out the processing by the transformer and pumps the value to the buffered channel
// to be read sequentially by the sync buffer go-routine.
func (m *messageBroadcast) pumpToOut(val string, t Transformer, send chan<- string, skip chan<- struct{}) func() error {
	f := func() error {
		// err should be ctx error if context cancelled.
		val, err := t.Transform(m.ctx, val)
		if err != nil {
			if err == ErrSkip {

				// skip value.
				select {
				case <-m.ctx.Done():
					return m.ctx.Err()
				case skip <- struct{}{}:
					return nil
				}
			}

			return err
		}

		// out could potentially be blocking.
		select {
		case <-m.ctx.Done():
			return m.ctx.Err()
		case send <- val:
			return nil
		}
	}

	return f
}

func (m *messageBroadcast) close() {
	for _, listener := range m.listeners {
		close(listener)
	}
}

func (l *UniformTransformerLayer) getStatefullTransformers() []Transformer {
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
