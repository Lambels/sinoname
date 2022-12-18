package sinoname

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"go.uber.org/goleak"
	"golang.org/x/sync/errgroup"
)

type timeoutTransformer struct {
	add string
	d   time.Duration
}

func (t timeoutTransformer) Transform(ctx context.Context, in MessagePacket) (MessagePacket, error) {
	select {
	case <-ctx.Done():
		return MessagePacket{}, ctx.Err()

	case <-time.After(t.d):
		in.setAndIncrement(in.Message + t.add)
		return in, nil
	}
}

func newTimeoutTransformer(add string, d time.Duration) TransformerFactory {
	return func(cfg *Config) (Transformer, bool) {
		return timeoutTransformer{add, d}, false
	}
}

type skipTransformer struct {
	add  string
	skip int
}

func (t skipTransformer) Transform(ctx context.Context, in MessagePacket) (MessagePacket, error) {
	in.Skip = t.skip
	in.setAndIncrement(in.Message + t.add)
	return in, nil
}

func newSkipTransformer(add string, skip int) TransformerFactory {
	return func(cfg *Config) (Transformer, bool) {
		return skipTransformer{add, skip}, false
	}
}

type errTransformer struct {
	err error
}

func (t errTransformer) Transform(context.Context, MessagePacket) (MessagePacket, error) {
	return MessagePacket{}, t.err
}

func newErrorTransformer(err error) TransformerFactory {
	return func(cfg *Config) (Transformer, bool) {
		return errTransformer{err}, false
	}
}

type statefullTransformer struct {
	state int32
}

func (t *statefullTransformer) Transform(ctx context.Context, in MessagePacket) (MessagePacket, error) {
	state := atomic.AddInt32(&t.state, 1)
	in.setAndIncrement(fmt.Sprintf("%v%d", in, state))
	return in, nil
}

func newStatefullTransformer() TransformerFactory {
	return func(cfg *Config) (Transformer, bool) {
		return &statefullTransformer{}, true
	}
}

type addTransformer struct {
	add string
}

func (t addTransformer) Transform(ctx context.Context, in MessagePacket) (MessagePacket, error) {
	in.setAndIncrement(in.Message + t.add)
	return in, nil
}

func newAddTransformer(add string) TransformerFactory {
	return func(cfg *Config) (Transformer, bool) {
		return addTransformer{add: add}, false
	}
}

func TestLayerCloseProducerChannel(t *testing.T) {
	defer goleak.VerifyNone(t)
	t.Run("Without_Values", func(t *testing.T) {
		layers := []Layer{
			newUniformLayer(Noop),
			newTransformerLayer(Noop),
		}

		for _, layer := range layers {
			ch := make(chan MessagePacket)
			close(ch)

			testLayerChanClose(t, context.Background(), layer, ch, 0, 1*time.Second, 1*time.Second)
		}
	})

	t.Run("With_Values", func(t *testing.T) {
		layers := []Layer{
			newUniformLayer(
				newTimeoutTransformer("1", 1*time.Second),
				newTimeoutTransformer("2", 2*time.Second),
			),
			newTransformerLayer(
				newTimeoutTransformer("1", 1*time.Second),
				newTimeoutTransformer("2", 2*time.Second),
			),
		}

		for _, layer := range layers {
			ch := make(chan MessagePacket, 1)
			ch <- MessagePacket{}
			close(ch)

			testLayerChanClose(t, context.Background(), layer, ch, 2, 3*time.Second, 1*time.Second)
		}
	})
}

func TestLayerContextCancel(t *testing.T) {
	defer goleak.VerifyNone(t)
	t.Run("Manual_Close", func(t *testing.T) {
		for _, v := range []struct {
			l                 Layer
			n                 int
			timeout, deadline time.Duration
		}{
			{
				l:        newUniformLayer(newTimeoutTransformer("1", 1*time.Microsecond), newTimeoutTransformer("2", 10*time.Second)),
				n:        0,
				timeout:  0,
				deadline: 3 * time.Second,
			},
			{
				l:        newTransformerLayer(newTimeoutTransformer("1", 1*time.Second), newTimeoutTransformer("2", 10*time.Second)),
				n:        1,
				timeout:  3 * time.Second,
				deadline: 2 * time.Second,
			},
		} {
			ch := make(chan MessagePacket, 1)
			ch <- MessagePacket{}

			ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
			testLayerChanClose(t, ctx, v.l, ch, v.n, v.timeout, v.deadline)
		}
	})

	t.Run("Error_Close", func(t *testing.T) {
		layers := []Layer{
			newUniformLayer(
				newErrorTransformer(errors.New("test error")),
				newTimeoutTransformer("1", 10*time.Second),
			),
			newTransformerLayer(
				newErrorTransformer(errors.New("test error")),
				newTimeoutTransformer("1", 10*time.Second),
			),
		}

		for _, layer := range layers {
			ch := make(chan MessagePacket, 1)
			ch <- MessagePacket{}

			testLayerChanClose(t, context.Background(), layer, ch, 0, 0, 1*time.Second)
		}
	})
}

func TestLayerSkip(t *testing.T) {
	defer goleak.VerifyNone(t)
	t.Run("Handle_Skip_Layers", func(t *testing.T) {
		layers := []Layers{
			[]Layer{
				newTransformerLayer(newSkipTransformer("1", 2)),
				newTransformerLayer(newAddTransformer("2")),
				newTransformerLayer(newAddTransformer("2")),
				newTransformerLayer(newAddTransformer("3")),
			},
			[]Layer{
				newUniformLayer(newSkipTransformer("1", 2)),
				newUniformLayer(newAddTransformer("2")),
				newUniformLayer(newAddTransformer("2")),
				newUniformLayer(newAddTransformer("3")),
			},
		}

		for _, l := range layers {
			recv, _, err := l.Run(context.Background(), MessagePacket{})
			if err != nil {
				t.Fatal(err)
			}

			v := <-recv
			if v.Message[len(v.Message)-1] != '3' || v.Message[len(v.Message)-2] == '2' {
				t.Fatal("didnt expect packet:", v)
			}
		}
	})
	t.Run("Handle_Skip_All_Layers", func(t *testing.T) {
		layers := []Layers{
			[]Layer{
				newTransformerLayer(newSkipTransformer("1", 3)),
				newTransformerLayer(newAddTransformer("2")),
				newTransformerLayer(newAddTransformer("2")),
				newTransformerLayer(newAddTransformer("2")),
			},
			[]Layer{
				newUniformLayer(newSkipTransformer("1", 3)),
				newUniformLayer(newAddTransformer("2")),
				newUniformLayer(newAddTransformer("2")),
				newUniformLayer(newAddTransformer("2")),
			},
		}

		for _, l := range layers {
			recv, _, err := l.Run(context.Background(), MessagePacket{})
			if err != nil {
				t.Fatal(err)
			}

			v := <-recv
			if len(v.Message) != 1 || v.Message[0] != '1' {
				t.Fatal("didnt expect packet:", v)
			}
		}
	})
}

func testLayerChanClose(t *testing.T, ctx context.Context, l Layer, src chan MessagePacket, n int, timeout, deadline time.Duration) {
	g, ctx := errgroup.WithContext(ctx)
	sink, err := l.PumpOut(ctx, g, src)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < n; i++ {
		select {
		case _, ok := <-sink:
			if !ok {
				t.Fatal("didnt expect sink to close yet")
			}
			// val recieced from sink.
		case <-time.After(timeout):
			t.Fatal("values should be available before timeout")
		}
	}

	select {
	case _, ok := <-sink:
		if ok {
			t.Fatal("recieved unexpected non close message")
		}
	case <-time.After(deadline):
		t.Fatal("expected channel to be closed", n)
	}
}
