package sinoname

import "context"

type numKey struct{}

// ContextWithNumber adds a number value to the context, this value will be linked to the
// value you are sending in the pipeline to generate more custom outcomes.
func ContextWithNumber(ctx context.Context, n int) context.Context {
	return context.WithValue(ctx, numKey{}, n)
}

// NumberFromContext gets the number value from the context.
func NumberFromContext(ctx context.Context) (int, bool) {
	n, ok := ctx.Value(numKey{}).(int)
	return n, ok
}

type stringKey struct{}

// ContextWithString adds a string value to the context, this value will be linked to the
// value you are sending in the pipeline to generate more custom outcomes.
func ContextWithString(ctx context.Context, v stringKey) context.Context {
	return context.WithValue(ctx, stringKey{}, v)
}

// StringFromContext gets the string value from the context.
func StringFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(stringKey{}).(string)
	return v, ok
}
