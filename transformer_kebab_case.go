package sinoname

import (
	"context"
	"strings"
)

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

	split := SplitOnSpecial(in, t.cfg.SplitOn)
	out := strings.Join(split, "-")
	if ok, err := t.cfg.Source.Valid(ctx, out); !ok || err != nil {
		return in, err
	}

	return out, nil
}
