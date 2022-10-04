package sinoname

import (
	"context"
	"strings"
)

var KebabCase = func(cfg *Config) Transformer {
	return &kebabCaseTransformer{
		maxLen:  cfg.MaxLen,
		source:  cfg.Source,
		special: cfg.Special,
	}
}

type kebabCaseTransformer struct {
	maxLen  int
	source  Source
	special []string
}

func (t *kebabCaseTransformer) Transform(ctx context.Context, in string) (string, error) {
	if len(in) > t.maxLen {
		return in, nil
	}

	split := SplitOnSpecial(in, t.special)
	out := strings.Join(split, "-")
	if ok, err := t.source.Valid(ctx, out); !ok || err != nil {
		return in, err
	}

	return out, nil
}
