package transformer

import "github.com/Lambels/sinoname/config"

var Noop = func(_ *config.Config) Transformer {
	return &noopTransformer{}
}

type noopTransformer struct{}

func (t *noopTransformer) Transform(in string) (string, error) {
	return in, nil
}
