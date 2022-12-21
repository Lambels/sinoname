package sinoname

import (
	"context"
	"strconv"
	"sync"
	"testing"
	"time"

	"go.uber.org/goleak"
	"golang.org/x/sync/errgroup"
)

type varSleepTransformer struct {
	root time.Duration
}

func (t varSleepTransformer) Transform(ctx context.Context, in MessagePacket) (MessagePacket, error) {
	div, err := strconv.Atoi(in.Message)
	if err != nil {
		return in, err
	}

	d := t.root / time.Duration(div)
	select {
	case <-ctx.Done():
		return in, ctx.Err()
	case <-time.After(d):
		return in, nil
	}
}

func TestBroadcasterOrder(t *testing.T) {
	defer goleak.VerifyNone(t)
	ch := make(chan MessagePacket, 5)
	ch <- MessagePacket{"1", 0, 0}
	ch <- MessagePacket{"2", 0, 0}
	ch <- MessagePacket{"3", 0, 0}
	ch <- MessagePacket{"4", 0, 0}
	ch <- MessagePacket{"5", 0, 0}
	close(ch)

	counter := 1
	var mu sync.Mutex
	wait, clnUp := notifyClose()

	out := newPacketBroadcatser(
		context.Background(),
		testConfig,
		ch,
		&errgroup.Group{},
		[]Transformer{varSleepTransformer{5 * time.Second}},
		func(_ context.Context, wg *sync.WaitGroup, _ int, v MessagePacket) error {
			mu.Lock()
			defer mu.Unlock()
			defer wg.Done()
			val, err := strconv.Atoi(v.Message)
			if err != nil {
				return err
			}

			if val != counter {
				t.Fatal("id doesent match current counter")
			}

			counter += 1
			return nil
		},
		nil,
		clnUp,
	)

	out.StartListen()
	select {
	case <-wait:
	case <-time.After(6 * time.Second):
		t.Fatal("expected to close before 6 seconds")
	}
}

func notifyClose() (chan struct{}, handlerExit) {
	ch := make(chan struct{})
	return ch, func(wg *sync.WaitGroup, b bool) {
		if b {
			close(ch)
			return
		}

		wg.Wait()
		close(ch)
	}
}
