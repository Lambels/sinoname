package layer

// syncOut waits for nWriters to write their value via the (*syncOut).write(),
// each writer should call the write function in its own go-routine.
type syncOut struct {
	nWriters int
	closeC   chan struct{}

	stateC chan *state
	outC   chan<- string
}

type state struct {
	waiters map[int]chan struct{}
	buf     []string
}

// flushAndNotify flushed all the buf values to out and notifies all the waiters.
//
// must be called with ownership to state.
func (s *state) flushAndNotify(to chan<- string, closeC chan struct{}) {
	for _, v := range s.buf {
		// potentially return faster.
		select {
		case <-closeC:
			return
		default:
		}

		select {
		case to <- v:
		case <-closeC: // free the state as quick as possible in the caller by returning early.
			return
		}
	}
	s.buf = nil

	for id, waiter := range s.waiters {
		close(waiter)
		delete(s.waiters, id)
	}
}

func newSyncOut(n int, out chan<- string) *syncOut {
	b := &syncOut{
		nWriters: n,
		closeC:   make(chan struct{}),

		stateC: make(chan *state, 1),
		outC:   out,
	}

	s := &state{
		waiters: make(map[int]chan struct{}),
		buf:     make([]string, 0),
	}

	b.stateC <- s
	return b
}

// close closes the out channel and makes write calls stop blocking and no-op.
func (b *syncOut) close() {
	close(b.closeC)
	<-b.stateC
	close(b.outC)
}

// write writes one value to the buf and then waits for the other write calls to write their
// value then unblocks.
func (b *syncOut) write(id int, val string) bool {
	select {
	case state := <-b.stateC:
		// if there is already a waiter for this id, wait.
		_, ok := state.waiters[id]
		if ok {
			b.stateC <- state
			return false
		}

		// write to state buffer.
		state.buf = append(state.buf, val)
		// last writer, no need to block.
		if len(state.buf) == b.nWriters {
			state.flushAndNotify(b.outC, b.closeC)
			b.stateC <- state
			return true
		}

		wait := make(chan struct{})
		state.waiters[id] = wait
		b.stateC <- state

		select {
		case <-wait:
			return true
		case <-b.closeC:
			return false
		}

	case <-b.closeC:
		return false
	}
}
