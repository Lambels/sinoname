package sinoname

import (
	"context"
	"strings"
)

// CamelCase splits on special characters and joins them back with capitalization.
//
// foo.bar_buz -> fooBarBuz
var CamelCase = func(cfg *Config) (Transformer, bool) {
	return &camelCaseTransformer{
		cfg: cfg,
	}, false
}

type camelCaseTransformer struct {
	cfg *Config
}

func (t *camelCaseTransformer) Transform(ctx context.Context, in string) (string, error) {
	if len(in) > t.cfg.MaxLen {
		return in, nil
	}

	split := t.cfg.Tokenize(in)
	for i := 1; i < len(split); i++ {
		split[i] = ucCapitalFirst(split[i])
	}

	out := strings.Join(split, "")
	if ok, err := t.cfg.Source.Valid(ctx, out); !ok || err != nil {
		return in, err
	}

	return out, nil
}
