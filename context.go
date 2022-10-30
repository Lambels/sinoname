package sinoname

import "context"

type numKey struct{}

func ContextWithNumber(ctx context.Context, n int) context.Context {
	return context.WithValue(ctx, numKey{}, n)
}

func NumberFromContext(ctx context.Context) (int, bool) {
	n, ok := ctx.Value(numKey{}).(int)
	return n, ok
}

type stringKey struct{}

func ContextWithString(ctx context.Context, v stringKey) context.Context {
	return context.WithValue(ctx, stringKey{}, v)
}

func StringFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(stringKey{}).(string)
	return v, ok
}

type sinkKey struct{}

func SinkFromContext(ctx context.Context) (chan<- string, bool) {
	v, ok := ctx.Value(sinkKey{}).(chan string)
	return v, ok
}
