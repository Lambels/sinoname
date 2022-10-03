package layer

import (
	"context"
	"testing"
	"time"

	"github.com/Lambels/sinoname/transformer"
	"golang.org/x/sync/errgroup"
)

func TestUnfiformLayerCloseProducerChannel(t *testing.T) {
	t.Parallel()
	t.Run("Without Values", func(t *testing.T) {
		t.Parallel()
		layer := newUniformLayer(transformer.Noop(nil))

		producer := make(chan string)
		sink, err := layer.PumpOut(context.Background(), &errgroup.Group{}, producer)
		if err != nil {
			t.Fatal(err)
		}

		close(producer)
		_, ok := <-sink
		if ok {
			t.Fatal("expected ok to be false")
		}
	})

	t.Run("With Values", func(t *testing.T) {
		t.Parallel()
		layer := newUniformLayer(
			timeoutTransformer{add: "1", d: 1 * time.Second},
			timeoutTransformer{add: "2", d: 2 * time.Second},
		)

		producer := make(chan string, 1)
		producer <- ""
		close(producer)

		sink, err := layer.PumpOut(context.Background(), &errgroup.Group{}, producer)
		if err != nil {
			t.Fatal(err)
		}

		for i := 0; i < 3; i++ {
			if i == 2 {
				select {
				case _, ok := <-sink:
					if ok {
						t.Fatal("expected channel to be closed after messages read")
					}
				case <-time.After(20 * time.Microsecond):
					t.Fatal("channel should be closed imediately")
				}
				return
			}

			select {
			case <-sink:
				// values still recieved even if producer closed.
			case <-time.After(3 * time.Second):
				t.Fatal("values should be available at around 2 seconds")
			}
		}
	})
}

func TestUniformLayerCloseCtx(t *testing.T) {
	t.Run("Manual", func(t *testing.T) {
		t.Parallel()
		layer := newUniformLayer(
			timeoutTransformer{add: "1", d: 1 * time.Microsecond},
			timeoutTransformer{add: "2", d: 10 * time.Second},
		)

		producer := make(chan string, 2)
		producer <- ""
		producer <- ""

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		sink, err := layer.PumpOut(ctx, &errgroup.Group{}, producer)
		if err != nil {
			t.Fatal(err)
		}

		<-ctx.Done()

		select {
		case _, ok := <-sink:
			if ok {
				t.Fatal("recieved unexpected non close message")
			}
		case <-time.After(2 * time.Second):
			t.Fatal("expected sink to be closed")
		}
	})

	t.Run("Error", func(t *testing.T) {
		t.Parallel()
		layer := newUniformLayer(
			errTransformer{},
			timeoutTransformer{add: "1", d: 1 * time.Second},
		)

		producer := make(chan string, 1)
		producer <- ""

		g, errCtx := errgroup.WithContext(context.Background())
		sink, err := layer.PumpOut(errCtx, g, producer)
		if err != nil {
			t.Fatal(err)
		}

		select {
		case _, ok := <-sink:
			if ok {
				t.Fatal("recieved unexpected non close message")
			}
		case <-time.After(1 * time.Second):
			t.Fatal("expected sink to be closed")
		}
	})
}

func TestUniformBatch(t *testing.T) {
	t.Parallel()

	producer := make(chan string, 3)
	producer <- ""
	producer <- ""
	producer <- ""
	close(producer)

	layer := newUniformLayer(
		timeoutTransformer{add: "1", d: 1 * time.Microsecond},
		timeoutTransformer{add: "2", d: 1 * time.Second},
	)
	sink, err := layer.PumpOut(context.Background(), &errgroup.Group{}, producer)
	if err != nil {
		t.Fatal(err)
	}

	for {
		<-time.After(2 * time.Second)
		if len(sink) != 0 && cap(sink) != len(sink) {
			t.Fatal("every two seconds the channel must be full")
		}

		v1, ok := <-sink
		if !ok {
			return
		}
		v2 := <-sink

		if v1 == v2 {
			t.Fatal("recieved values must be different")
		}
	}

}

func newUniformLayer(t ...transformer.Transformer) *UniformTransformerLayer {
	layer := UniformTransformerLayer{
		Transformers: make([]transformer.Transformer, 0),
	}

	layer.Transformers = append(layer.Transformers, t...)
	return &layer
}
