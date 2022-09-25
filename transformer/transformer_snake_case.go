package transformer

import (
	"context"
	"strings"

	"github.com/Lambels/sinoname/config"
	"github.com/Lambels/sinoname/helper"
)

var SnakeCase = func(cfg *config.Config) Transformer {
	return &snakeCaseTransformer{
		maxLen: cfg.MaxLen,
		source: cfg.Source,
	}
}

type snakeCaseTransformer struct {
	maxLen int
	source config.Source
}

func (t *snakeCaseTransformer) Transform(ctx context.Context, in string) (string, error) {
	if len(in) > t.maxLen {
		return in, nil
	}

	split := helper.SplitOnSpecial(in)
	out := strings.Join(split, "_")
	if ok, err := t.source.Valid(ctx, out); !ok || err != nil {
		return in, err
	}

	return out, nil
}
