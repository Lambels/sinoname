package sinoname

import (
	"context"
)

var Plural = func(cfg *Config) (Transformer, bool) {
	return &pluralTransformer{
		maxLen: cfg.MaxLen,
		source: cfg.Source,
	}, false
}

type pluralTransformer struct {
	maxLen int
	source Source
}

func (t *pluralTransformer) Transform(ctx context.Context, in string) (string, error) {
	out := in + "s"
	if len(out) > t.maxLen {
		return in, nil
	}

	if ok, err := t.source.Valid(ctx, out); !ok || err != nil {
		return in, err
	}

	return out, nil
}
