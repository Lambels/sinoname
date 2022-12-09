package sinoname

import (
	"context"
	"sort"
	"sync"

	"golang.org/x/sync/errgroup"
)

// handler handels a message value from a transformer.
// responsible for decrementing the waitgroup.
type handlerValue func(ctx context.Context, wg *sync.WaitGroup, id int, v MessagePacket) error

type handlerExit func(*sync.WaitGroup, bool)

type waiter struct {
	idT, idV int
	skipped  bool
	value    MessagePacket
}

type packetBroadcaster struct {
	src <-chan MessagePacket
	ctx context.Context

	wg     *sync.WaitGroup
	runner *errgroup.Group

	transformers []Transformer
	receive      []chan *waiter

	valuesCount int

	handleValue handlerValue
	handleSkip  handlerValue
	handleExit  handlerExit
}

func newPacketBroadcatser(ctx context.Context, src <-chan MessagePacket, g *errgroup.Group, t []Transformer, handleValue, handleSkip handlerValue, handleExit handlerExit) *packetBroadcaster {
	return &packetBroadcaster{
		src:         src,
		wg:          &sync.WaitGroup{},
		runner:      g,
		handleValue: handleValue,
		handleExit:  handleExit,
	}
}

func (b *packetBroadcaster) StartListen() {
	for _, ch := range b.receive {
		b.runner.Go(
			b.recieveValues(ch),
		)
	}

	go b.listen()
}

func (b *packetBroadcaster) listen() {
	for {
		select {
		case <-b.ctx.Done():
			b.exit(true)
			return

		case v, ok := <-b.src:
			if !ok {
				b.exit(false)
				return
			}

			if v.Skip > 0 {
				v.Skip--
				b.handleSkip(b.ctx, nil, -1, v)
				continue
			}

			idV := b.valuesCount
			b.wg.Add(len(b.transformers))
			for i, t := range b.transformers {
				b.runner.Go(b.runTransformer(t, v, i, idV))
			}
			b.valuesCount++
		}
	}
}

func (b *packetBroadcaster) exit(cancel bool) {
	b.handleExit(b.wg, cancel)
	b.wg.Wait()
	for _, ch := range b.receive {
		close(ch)
	}
}

// runTransformer runs the transformer and sends the value to the waiter queue.
func (b *packetBroadcaster) runTransformer(t Transformer, v MessagePacket, idT, idV int) func() error {
	return func() error {
		out, err := t.Transform(b.ctx, v)
		w := &waiter{
			idT:   idT,
			idV:   idV,
			value: out,
		}
		if err != nil {
			if err != ErrSkip {
				return err
			}
			w.skipped = true
		}

		ch := b.receive[idT]
		select {
		case <-b.ctx.Done():
			return b.ctx.Err()
		case ch <- w:
			return nil
		}
	}
}

// byIdV sorts the waiter by idV in ascending order.
type byIdV []*waiter

func (s byIdV) Len() int           { return len(s) }
func (s byIdV) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s byIdV) Less(i, j int) bool { return s[i].idV < s[j].idV }

// recieveValues recieves values from the waiter queue in order.
func (b *packetBroadcaster) recieveValues(ch <-chan *waiter) func() error {
	return func() error {
		var localIdV int
		waiterBuf := make([]*waiter, 0)

		for {
			// first check if we can process any values with what we currently have.
			sort.Sort(byIdV(waiterBuf))
			for i, w := range waiterBuf {
				if w.idV > localIdV {
					break
				}

				err := b.runWaiter(w)
				if err != nil {
					return err
				}
				// shrink via reslicing (array wont grow allot)
				waiterBuf = waiterBuf[i:]
			}

			select {
			case <-b.ctx.Done():
				// values which will never run but still occupy the wg.
				for range waiterBuf {
					b.wg.Done()
				}

				return b.ctx.Err()
			case w, ok := <-ch:
				if !ok {
					return nil
				}

				switch {
				case w.idV < localIdV: // run now but dont increment idV.
					return b.runWaiter(w)

				case w.idV == localIdV: // run now and increment idV.
					localIdV++
					return b.runWaiter(w)

				default: // waiter which will run in the future.
					waiterBuf = append(waiterBuf, w)
				}
			}

		}
	}
}

func (b *packetBroadcaster) runWaiter(w *waiter) error {
	if w.skipped {
		return b.handleSkip(b.ctx, b.wg, w.idT, w.value)
	}
	return b.handleValue(b.ctx, b.wg, w.idT, w.value)
}
