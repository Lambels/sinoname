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
	if len(in)+1 > t.maxLen {
		return in, nil
	}
	out := in + "s"

	if ok, err := t.source.Valid(ctx, out); !ok || err != nil {
		return in, err
	}

	return out, nil
}
