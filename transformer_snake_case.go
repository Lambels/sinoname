package sinoname

import (
	"context"
	"strings"
)

var SnakeCase = func(cfg *Config) (Transformer, bool) {
	return &snakeCaseTransformer{
		cfg: cfg,
	}, false
}

type snakeCaseTransformer struct {
	cfg *Config
}

func (t *snakeCaseTransformer) Transform(ctx context.Context, in string) (string, error) {
	if len(in) > t.cfg.MaxLen {
		return in, nil
	}

	split := SplitOnSpecial(in, t.cfg.SplitOn)
	out := strings.Join(split, "_")
	if ok, err := t.cfg.Source.Valid(ctx, out); !ok || err != nil {
		return in, err
	}

	return out, nil
}
