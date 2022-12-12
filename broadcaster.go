package sinoname

import (
	"context"
	"sort"
	"sync"

	"golang.org/x/sync/errgroup"
)

// handlerValue handels a value from the broadcaster.
type handlerValue func(ctx context.Context, wg *sync.WaitGroup, id int, v MessagePacket) error

// handels the exit from the transformer. If true, the exit was caused by a context cancel.
type handlerExit func(*sync.WaitGroup, bool)

// waiter indicates a packet waiting to be handeled.
type waiter struct {
	// idT represents the transformers index in the transformer slice.
	idT int
	// idV represents the id of the packet.
	// It is used to reconstruct the stream on the recieving end (processValues go-routine)
	// in such way that the results go out in the same order they come in.
	idV int
	// skipped indicates wether the packet was skipped.
	skipped bool
	// packet copy.
	value MessagePacket
}

// packetBroadcaster broadcasts packages to the transformers.
type packetBroadcaster struct {
	src <-chan MessagePacket
	ctx context.Context

	// lWg is a local waitgroup used to monitor how many transformer workers are
	// currently writing to their wait queues.
	//
	// Two separate waitgroups are used to eliminate the deadlock which would
	// occur when the processing go routine would wait for an exit signal to
	// process its last value and the exit signal would wait for the processing
	// go routine to finish off its last value for it to be sent.
	lWg sync.WaitGroup
	// pWg is a waitgroup used to monitor the values which are and will be processed.
	pWg *sync.WaitGroup
	// used to run go-routines: processValues and runTransformer.
	runner *errgroup.Group

	transformers []Transformer
	// the wait queue for waiters.
	receive []chan *waiter

	// valuesCount is used to give
	valuesCount int

	// handels a value from the transformer.
	handleValue handlerValue
	// handels a skipped value either from a transformer (via ErrSkip) or from a
	// complete layer skip (via the packet Skip field).
	//
	// If the packet is skipped via a transformer the id argument will be >= 0 .
	//
	// If the packet has skipped the whole layer the id argument will be = -1.
	handleSkip handlerValue
	handleExit handlerExit
}

func newPacketBroadcatser(ctx context.Context, src <-chan MessagePacket, g *errgroup.Group, t []Transformer, handleValue, handleSkip handlerValue, handleExit handlerExit) *packetBroadcaster {
	recievers := make([]chan *waiter, len(t))
	for i := range recievers {
		recievers[i] = make(chan *waiter)
	}

	return &packetBroadcaster{
		src: src,
		ctx: ctx,

		lWg:    sync.WaitGroup{},
		pWg:    &sync.WaitGroup{},
		runner: g,

		transformers: t,
		receive:      recievers,

		handleValue: handleValue,
		handleSkip:  handleSkip,
		handleExit:  handleExit,
	}
}

// StartListen starts the packet broadcaster in a separate go routine,
// reading values from the src channel.
func (b *packetBroadcaster) StartListen() {
	for _, ch := range b.receive {
		b.runner.Go(
			b.processValues(ch),
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
			b.lWg.Add(len(b.transformers))
			for i, t := range b.transformers {
				b.runner.Go(b.runTransformer(t, v, i, idV))
			}
			b.valuesCount++
		}
	}
}

func (b *packetBroadcaster) exit(cancel bool) {
	// wait for all values to be writted to their specific recieve channels.
	b.lWg.Wait()
	for _, ch := range b.receive {
		close(ch)
	}
	b.handleExit(b.pWg, cancel)
}

// runTransformer runs the transformer with the provided value.
//
// If the transformation is sucessful the value ends up in a wait queue to be ran by
// the value processor go routine. (waitgroup responsability shifted to value processor go routine)
//
// If the context is cancelled the routine exits and the waitgroup is decremented cleaning everything up.
//
// If the transformer returns an error, simillarly routine exits and the waitgroup is decremented.
// When the routine exits it also returns the transformer error cancelling the context.
func (b *packetBroadcaster) runTransformer(t Transformer, v MessagePacket, idT, idV int) func() error {
	return func() error {
		defer b.lWg.Done()
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
		b.pWg.Add(1) // shift wg responsability to processor.
		select {
		case <-b.ctx.Done():
			b.pWg.Done() // decrement processor wg since it will never reach it.
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

// processValues processes the wait queue for the specific transformer
func (b *packetBroadcaster) processValues(ch <-chan *waiter) func() error {
	return func() (err error) {
		var localIdV int
		waiterBuf := make([]*waiter, 0)

		defer func() {
			// if exciting with error just decrement pWg since values will be unsignificant.
			if err != nil {
				for range waiterBuf {
					b.pWg.Done()
				}
				return
			}

			sort.Sort(byIdV(waiterBuf))
			for _, w := range waiterBuf {
				if err = b.runWaiter(w); err != nil {
					return
				}
			}
		}()

		for {
			// first check if we can process any values with what we currently have.
			sort.Sort(byIdV(waiterBuf))
			for i, w := range waiterBuf {
				if w.idV > localIdV {
					break
				}

				err = b.runWaiter(w)
				if err != nil {
					return err
				}
				// shrink via reslicing (array wont grow allot)
				waiterBuf = waiterBuf[i+1:]
			}

			select {
			case <-b.ctx.Done():
				err = b.ctx.Err()
				return err
			case w, ok := <-ch:
				if !ok {
					return nil
				}

				switch {
				case w.idV < localIdV: // run now but dont increment idV.
					if err = b.runWaiter(w); err != nil {
						return err
					}

				case w.idV == localIdV: // run now and increment idV.
					localIdV++
					if err = b.runWaiter(w); err != nil {
						return err
					}

				default: // waiter which will run in the future.
					waiterBuf = append(waiterBuf, w)
				}
			}
		}
	}
}

func (b *packetBroadcaster) runWaiter(w *waiter) error {
	if w.skipped {
		return b.handleSkip(b.ctx, b.pWg, w.idT, w.value)
	}
	return b.handleValue(b.ctx, b.pWg, w.idT, w.value)
}
