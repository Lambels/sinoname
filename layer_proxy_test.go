package sinoname

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"
)

type noopAcceptingProxy struct{}

func (n noopAcceptingProxy) Proxy(context.Context, string) error {
	return nil
}

type noopSkipProxy struct{}

func (n noopSkipProxy) Proxy(context.Context, string) error {
	return errors.New("skip input")
}

type noopStopProxy struct{}

func (n noopStopProxy) Proxy(context.Context, string) error {
	return ErrQuit
}

func TestProxyLayerCloseProducerChannel(t *testing.T) {
	t.Parallel()
	t.Run("Without Values", func(t *testing.T) {
		t.Parallel()
		layer := newProxyLayer(noopAcceptingProxy{})

		producer := make(chan string)
		sink, err := layer.PumpOut(context.Background(), &errgroup.Group{}, producer)
		if err != nil {
			t.Fatal(err)
		}

		close(producer)

		select {
		case _, ok := <-sink:
			if ok {
				t.Fatal("expected ok to be false")
			}
		case <-time.After(20 * time.Microsecond):
			t.Fatal("expected channel to be closed")
		}
	})

	t.Run("With Values", func(t *testing.T) {
		t.Parallel()
		layer := newProxyLayer(noopAcceptingProxy{})

		producer := make(chan string, 1)
		producer <- "val"
		close(producer)

		sink, err := layer.PumpOut(context.Background(), &errgroup.Group{}, producer)
		if err != nil {
			t.Fatal(err)
		}

		select {
		case <-sink:
			// value recieved.
		case <-time.After(20 * time.Microsecond):
			t.Fatal("expected value")
		}

		select {
		case _, ok := <-sink:
			if ok {
				t.Fatal("expected closed channel")
			}
		case <-time.After(20 * time.Second):
			t.Fatal("expected closed channel")
		}
	})
}

type sleepAcceptingProxy struct {
	d time.Duration
}

func (p sleepAcceptingProxy) Proxy(ctx context.Context, _ string) error {
	select {
	case <-time.After(p.d):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func TestProxyLayerCancelContext(t *testing.T) {
	t.Parallel()
	t.Run("Manual", func(t *testing.T) {
		t.Parallel()
		layer := newProxyLayer(sleepAcceptingProxy{3 * time.Second})

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
		case <-time.After(1 * time.Second):
			t.Fatal("expected sink to be closed")
		}
	})

	t.Run("Stop Proxy", func(t *testing.T) {
		t.Parallel()
		layer := newProxyLayer(noopStopProxy{})

		producer := make(chan string, 1)
		producer <- ""

		g, ctx := errgroup.WithContext(context.Background())
		sink, err := layer.PumpOut(ctx, g, producer)
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

func TestProxyLayerSkip(t *testing.T) {
	t.Parallel()
	layer := newProxyLayer(noopAcceptingProxy{}, noopSkipProxy{})

	producer := make(chan string, 2)
	producer <- ""
	producer <- ""
	close(producer)

	sink, err := layer.PumpOut(context.Background(), &errgroup.Group{}, producer)
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
}

func newProxyLayer(p ...Proxy) *ProxyLayer {
	layer := ProxyLayer{
		proxys: make([]Proxy, 0),
	}

	layer.proxys = append(layer.proxys, p...)
	return &layer
}
