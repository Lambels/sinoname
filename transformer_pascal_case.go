package sinoname

import (
	"context"
	"strings"
)

var CamelCase = func(cfg *Config) (Transformer, bool) {
	return &camelCaseTransformer{
		cfg: cfg,
	}, false
}

type camelCaseTransformer struct {
	cfg *Config
}

func (t *camelCaseTransformer) Transform(ctx context.Context, in string) (string, error) {
	if len(in) > t.cfg.MaxLen {
		return in, nil
	}

	split := SplitOnSpecial(in, t.cfg.SplitOn)
	for i, word := range split {
		split[i] = ucCapitalFirst(word)
	}

	out := strings.Join(split, "")
	if ok, err := t.cfg.Source.Valid(ctx, out); !ok || err != nil {
		return in, err
	}

	return out, nil
}

func ucCapitalFirst(val string) string {
	cap := ucCapital(val)
	return cap + val[len(cap):]
}
