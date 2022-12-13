package sinoname

import (
	"context"
	"testing"
	"time"

	"go.uber.org/goleak"
	"golang.org/x/sync/errgroup"
)

func TestUniformBatch(t *testing.T) {
	defer goleak.VerifyNone(t)

	producer := make(chan MessagePacket, 3)
	producer <- MessagePacket{}
	producer <- MessagePacket{}
	producer <- MessagePacket{}
	close(producer)

	layer := newUniformLayer(
		newTimeoutTransformer("1", 1*time.Microsecond),
		newTimeoutTransformer("2", 1*time.Second),
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

func newUniformLayer(tf ...TransformerFactory) *UniformTransformerLayer {
	layer := &UniformTransformerLayer{
		transformers:         make([]Transformer, len(tf)),
		transformerFactories: make([]TransformerFactory, 0),
	}

	for i, f := range tf {
		t, statefull := f(testConfig)
		if statefull {
			layer.transformerFactories = append(layer.transformerFactories, f)
		}
		layer.transformers[i] = t
	}
	return layer
}
