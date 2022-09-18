package layer

import "sync"

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
