package sinoname

import (
	"context"
	"strings"
	"unicode"
)

var CamelCase = func(cfg *Config) Transformer {
	return &camelCaseTransformer{
		maxLen:  cfg.MaxLen,
		source:  cfg.Source,
		special: cfg.Special,
	}
}

type camelCaseTransformer struct {
	maxLen  int
	source  Source
	special []string
}

func (t *camelCaseTransformer) Transform(ctx context.Context, in string) (string, error) {
	if len(in) > t.maxLen {
		return in, nil
	}

	split := SplitOnSpecial(in, t.special)
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
