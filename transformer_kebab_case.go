package sinoname

import (
	"context"
	"strings"
)

// KebabCase splits on special characters and joins them back with "-".
//
// Foo.Bar_Buz -> Foo-Bar-Buz
var KebabCase = func(cfg *Config) (Transformer, bool) {
	return &kebabCaseTransformer{
		cfg: cfg,
	}, false
}

type kebabCaseTransformer struct {
	cfg *Config
}

func (t *kebabCaseTransformer) Transform(ctx context.Context, in string) (string, error) {
	if len(in) > t.cfg.MaxLen {
		return in, nil
	}

	split := SplitOnSpecial(in)
	out := strings.Join(split, "-")
	if ok, err := t.cfg.Source.Valid(ctx, out); !ok || err != nil {
		return in, err
	}

	return out, nil
}
