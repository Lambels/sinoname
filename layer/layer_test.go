package layer

import (
	"context"
	"errors"
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

type errTransformer struct{}

func (t errTransformer) Transform(context.Context, string) (string, error) {
	return "", errors.New("test error")
}
