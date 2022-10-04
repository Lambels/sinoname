package sinoname

import (
	"context"
	"strings"
)

var SnakeCase = func(cfg *Config) Transformer {
	return &snakeCaseTransformer{
		maxLen: cfg.MaxLen,
		source: cfg.Source,
	}
}

type snakeCaseTransformer struct {
	maxLen  int
	source  Source
	special []string
}

func (t *snakeCaseTransformer) Transform(ctx context.Context, in string) (string, error) {
	if len(in) > t.maxLen {
		return in, nil
	}

	split := SplitOnSpecial(in, t.special)
	out := strings.Join(split, "_")
	if ok, err := t.source.Valid(ctx, out); !ok || err != nil {
		return in, err
	}

	return out, nil
}
