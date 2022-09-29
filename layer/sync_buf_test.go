package layer

import (
	"testing"
	"time"
)

func TestSyncBufWrite(t *testing.T) {
	t.Run("1_writer", func(t *testing.T) {
		syncBuf := newSyncBuf(1, make(chan string, 1))

		select {
		case <-writeNotify(syncBuf, "val"):

		case <-time.After(20 * time.Millisecond):
			t.Fatal("write with single writer shouldnt block")
		}
	})

	t.Run("2_writers", func(t *testing.T) {
		out := make(chan string, 2)
		syncBuf := newSyncBuf(2, out)

		go func() {
			time.Sleep(2 * time.Second)
			syncBuf.write("val2")
		}()

		select {
		case <-writeNotify(syncBuf, "val1"):
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
	syncBuf := newSyncBuf(2, out)

	go func() {
		time.Sleep(2 * time.Second)
		syncBuf.close()
	}()

	select {
	case <-writeNotify(syncBuf, "val1"):
		select {
		case <-out:
			t.Fatal("expected no value")

		default:
		}

	case <-time.After(3 * time.Second):
		t.Fatal("write notify shouldv triggered withing 3 seconds")
	}
}

func writeNotify(buf *syncBuffer, val string) <-chan struct{} {
	ch := make(chan struct{}, 1)
	go func() {
		buf.write(val)
		ch <- struct{}{}
	}()

	return ch
}
