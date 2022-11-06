package sinoname

import (
	"context"
)

var Plural = func(cfg *Config) (Transformer, bool) {
	return &pluralTransformer{
		cfg: cfg,
	}, false
}

type pluralTransformer struct {
	cfg *Config
}

func (t *pluralTransformer) Transform(ctx context.Context, in string) (string, error) {
	if len(in)+1 > t.cfg.MaxLen {
		return in, nil
	}
	out := in + "s"

	if ok, err := t.cfg.Source.Valid(ctx, out); !ok || err != nil {
		return in, err
	}

	return out, nil
}
