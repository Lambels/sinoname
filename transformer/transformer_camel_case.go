package transformer

import (
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

func (t *camelCaseTransformer) Transform(in string) (string, error) {
	if len(in) > t.maxLen {
		return in, nil
	}

	split := helper.SplitOnSpecial(in)
	words := strings.Fields(split)
	for i, word := range words {
		words[i] = ucCapitalFirst(word)
	}

	out := strings.Join(words, "")
	if ok, err := t.source.Valid(out); !ok || err != nil {
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