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

func (t timeoutTransformer) Transform(ctx context.Context, val MessagePacket) (MessagePacket, error) {
	select {
	case <-ctx.Done():
		return MessagePacket{}, ctx.Err()

	case <-time.After(t.d):
		val.Message = val.Message + t.add
		return val, nil
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

func (t skipTransformer) Transform(ctx context.Context, val MessagePacket) (MessagePacket, error) {
	val.Skip = t.skip
	val.Message += t.add
	return val, nil
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
	in.Message = fmt.Sprintf("%v:%d", in, state)
	return in, nil
}

func newStatefullTransformer() TransformerFactory {
	return func(cfg *Config) (Transformer, bool) {
		return &statefullTransformer{}, true
	}
}
