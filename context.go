package sinoname

import "context"

type sinkKey struct{}

func SinkFromContext(ctx context.Context) (chan<- string, bool) {
	v, ok := ctx.Value(sinkKey{}).(chan string)
	return v, ok
}
