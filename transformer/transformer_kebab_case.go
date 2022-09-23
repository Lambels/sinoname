package transformer

import (
	"context"

	"github.com/Lambels/sinoname/config"
	"github.com/Lambels/sinoname/helper"
)

var kebabCaseMap map[rune][]rune = map[rune][]rune{
	' ': {'-'},
}

var KebabCase = func(cfg *config.Config) Transformer {
	return &kebabCaseTransformer{
		maxLen: cfg.MaxLen,
		source: cfg.Source,
	}
}

type kebabCaseTransformer struct {
	maxLen int
	source config.Source
}

func (t *kebabCaseTransformer) Transform(ctx context.Context, in string) (string, error) {
	if len(in) > t.maxLen {
		return in, nil
	}

	split := helper.SplitOnSpecial(in)
	out := helper.ReplaceRunes(split, kebabCaseMap)
	if ok, err := t.source.Valid(ctx, out); !ok || err != nil {
		return in, err
	}

	return out, nil
}
