package sinoname

import (
	"testing"
	"time"
)

func TestSyncBufWrite(t *testing.T) {
	t.Run("1_writer", func(t *testing.T) {
		syncBuf := newSyncOut(1, make(chan string, 1))

		select {
		case <-writeNotify(syncBuf, 1, "val"):

		case <-time.After(20 * time.Millisecond):
			t.Fatal("write with single writer shouldnt block")
		}
	})

	t.Run("2_writers", func(t *testing.T) {
		out := make(chan string, 2)
		syncBuf := newSyncOut(2, out)

		go func() {
			time.Sleep(2 * time.Second)
			syncBuf.write(1, "val2")
		}()

		select {
		case <-writeNotify(syncBuf, 2, "val1"):
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
	out := make(chan string, 2)
	syncBuf := newSyncOut(2, out)

	go func() {
		time.Sleep(1 * time.Second)
		syncBuf.close()
	}()

	select {
	case <-writeNotify(syncBuf, 1, "val1"):
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

func writeNotify(buf *syncOut, id int, val string) <-chan struct{} {
	ch := make(chan struct{}, 1)
	go func() {
		buf.write(id, val)
		ch <- struct{}{}
	}()

	return ch
}
