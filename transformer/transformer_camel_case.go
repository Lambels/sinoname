package transformer

import (
	"context"
	"strings"
	"unicode"

	"github.com/Lambels/sinoname/config"
	"github.com/Lambels/sinoname/helper"
)

var CamelCase = func(cfg *config.Config) Transformer {
	return &camelCaseTransformer{
		maxLen: cfg.MaxLen,
		source: cfg.Source,
	}
}

type camelCaseTransformer struct {
	maxLen int
	source config.Source
}

func (t *camelCaseTransformer) Transform(ctx context.Context, in string) (string, error) {
	if len(in) > t.maxLen {
		return in, nil
	}

	split := helper.SplitOnSpecial(in)
	for i, word := range split {
		split[i] = ucCapitalFirst(word)
	}

	out := strings.Join(split, "")
	if ok, err := t.source.Valid(ctx, out); !ok || err != nil {
		return in, err
	}

	return out, nil
}

func ucCapitalFirst(val string) string {
	for _, v := range val {
		return string(unicode.ToUpper(v)) + val[len(string(v)):]
	}
	return ""
}
