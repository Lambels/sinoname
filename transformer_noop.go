package sinoname

import (
	"context"
)

var Noop = func(_ *Config) Transformer {
	return &noopTransformer{}
}

type noopTransformer struct{}

func (t *noopTransformer) Transform(_ context.Context, in string) (string, error) {
	return in, nil
}
