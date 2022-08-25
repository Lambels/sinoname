package sinoname

import (
	"strings"
	"unicode"
)

var CamelCase = func(cfg *Config) *Layer {
	layer := &Layer{
		cfg:          cfg,
		transformers: make([]Transformer, 1),
	}
	layer.transformers[0] = &camelCaseTransformer{
		maxLen: cfg.MaxLen,
		source: cfg.Source,
	}
	return layer
}

type camelCaseTransformer struct {
	maxLen int
	source Source
}

func (t *camelCaseTransformer) Transform(in string) (string, error) {
	if len(in) > t.maxLen {
		return in, nil
	}

	split := splitOnSpecial(in)
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
