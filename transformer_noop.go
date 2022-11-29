package sinoname

import (
	"context"
)

// Noop as the name says, doesent modify the incoming string.
var Noop = func(_ *Config) (Transformer, bool) {
	return &noopTransformer{}, false
}

type noopTransformer struct{}

func (t *noopTransformer) Transform(_ context.Context, in string) (string, error) {
	return in, nil
}
