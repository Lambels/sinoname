package sinoname

import (
	"testing"
	"time"
)

func TestSyncBufWrite(t *testing.T) {
	t.Run("1_writer", func(t *testing.T) {
		syncBuf := newSyncOut(1, make(chan MessagePacket, 1))

		select {
		case <-writeNotify(syncBuf, 1, MessagePacket{"val", 0, 0}):

		case <-time.After(20 * time.Millisecond):
			t.Fatal("write with single writer shouldnt block")
		}
	})

	t.Run("2_writers", func(t *testing.T) {
		out := make(chan MessagePacket, 2)
		syncBuf := newSyncOut(2, out)

		go func() {
			time.Sleep(2 * time.Second)
			syncBuf.Write(1, MessagePacket{"val2", 0, 0})
		}()

		select {
		case <-writeNotify(syncBuf, 2, MessagePacket{"val1", 0, 0}):
			v1, v2 := <-out, <-out
			if v1 == v2 {
				t.Fatal("got same value")
			}

		case <-time.After(3 * time.Second):
			t.Fatal("write notify shouldv triggered withing 3 seconds")
		}
	})
}

func TestSyncBufClose(t *testing.T) {
	out := make(chan MessagePacket, 2)
	syncBuf := newSyncOut(2, out)

	go func() {
		time.Sleep(1 * time.Second)
		syncBuf.Close()
	}()

	select {
	case <-writeNotify(syncBuf, 1, MessagePacket{"val1", 0, 0}):
		select {
		case _, ok := <-out:
			if ok {
				t.Fatal("expected a closed channel")
			}
		default:
			t.Fatal("expected channel to get closed imediatley after writer exit")
		}

	case <-time.After(2 * time.Second):
		t.Fatal("write notify shouldv triggered withing 3 seconds")
	}
}

func writeNotify(buf *syncOut, id int, val MessagePacket) <-chan struct{} {
	ch := make(chan struct{}, 1)
	go func() {
		buf.Write(id, val)
		ch <- struct{}{}
	}()

	return ch
}
