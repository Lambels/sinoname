package sinoname

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"
)

type timeoutTransformer struct {
	add string
	d   time.Duration
}

func (t timeoutTransformer) Transform(ctx context.Context, val string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()

	case <-time.After(t.d):
		return val + t.add, nil
	}
}

func newTimeoutTransformer(add string, d time.Duration) TransformerFactory {
	return func(cfg *Config) (Transformer, bool) {
		return timeoutTransformer{add, d}, false
	}
}

type shortcutTransformer struct {
	add string
}

func (t shortcutTransformer) Transform(ctx context.Context, in string) (string, error) {
	sink, _ := SinkFromContext(ctx)

	sink <- in + t.add
	return "", ErrSkip
}

func newShortcutTransformer(add string) TransformerFactory {
	return func(cfg *Config) (Transformer, bool) {
		return shortcutTransformer{add}, false
	}
}

type errTransformer struct {
	err error
}

func (t errTransformer) Transform(context.Context, string) (string, error) {
	return "", t.err
}

func newErrorTransformer(err error) TransformerFactory {
	return func(cfg *Config) (Transformer, bool) {
		return errTransformer{err}, false
	}
}

type statefullTransformer struct {
	state int32
}

func (t *statefullTransformer) Transform(ctx context.Context, in string) (string, error) {
	state := atomic.AddInt32(&t.state, 1)
	return fmt.Sprintf("%v:%d", in, state), nil
}

func newStatefullTransformer() TransformerFactory {
	return func(cfg *Config) (Transformer, bool) {
		return &statefullTransformer{}, true
	}
}
