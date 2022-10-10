package sinoname

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
	n       int
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
	s.n = 0

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

// wait starts waiting on id.
func (b *syncOut) wait(s *state, id int) bool {
	wait := make(chan struct{})
	s.waiters[id] = wait
	b.stateC <- s

	select {
	case <-wait:
		return true
	case <-b.closeC:
		return false
	}
}

// advance advances the writer without writing any actuall value.
func (b *syncOut) advance(id int) bool {
	select {
	case state := <-b.stateC:
		// if there is already a waiter for this id, void this entry.
		_, ok := state.waiters[id]
		if ok {
			b.stateC <- state
			return false
		}

		state.n++
		// last writer, no need to block.
		if state.n == b.nWriters {
			state.flushAndNotify(b.outC, b.closeC)
			b.stateC <- state
			return true
		}

		// start waiting.
		return b.wait(state, id)

	case <-b.closeC:
		return false
	}
}

// write writes one value to the buf and then waits for the other write calls to write their
// value then unblocks.
func (b *syncOut) write(id int, val string) bool {
	select {
	case state := <-b.stateC:
		// if there is already a waiter for this id, void this entry.
		_, ok := state.waiters[id]
		if ok {
			b.stateC <- state
			return false
		}

		// write to state buffer.
		state.buf = append(state.buf, val)
		state.n++
		// last writer, no need to block.
		if state.n == b.nWriters {
			state.flushAndNotify(b.outC, b.closeC)
			b.stateC <- state
			return true
		}

		// start waiting.
		return b.wait(state, id)

	case <-b.closeC:
		return false
	}
}
