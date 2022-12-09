package sinoname

import (
	"context"
	"strings"
)

// PascalCase splits on special characters and joins them back with capitalization.
//
// foo.bar_buz -> FooBarBuz
var PascalCase = func(cfg *Config) (Transformer, bool) {
	return &pascalCaseTransformer{
		cfg: cfg,
	}, false
}

type pascalCaseTransformer struct {
	cfg *Config
}

func (t *pascalCaseTransformer) Transform(ctx context.Context, in MessagePacket) (MessagePacket, error) {
	split := t.cfg.Tokenize(in.Message)
	for i, word := range split {
		split[i] = ucCapitalFirst(word)
	}

	out := strings.Join(split, "")
	if ok, err := t.cfg.Source.Valid(ctx, out); !ok || err != nil {
		return in, err
	}

	in.setAndIncrement(out)
	return in, nil
}

func ucCapitalFirst(val string) string {
	cap := ucCapital(val)
	return cap + val[len(cap):]
}
