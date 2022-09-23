package transformer

import (
	"context"

	"github.com/Lambels/sinoname/config"
)

var Noop = func(_ *config.Config) Transformer {
	return &noopTransformer{}
}

type noopTransformer struct{}

func (t *noopTransformer) Transform(_ context.Context, in string) (string, error) {
	return in, nil
}
