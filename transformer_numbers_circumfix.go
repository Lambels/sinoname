package sinoname

import (
	"context"
	"strconv"
)

var NumbersCircumfix = func(sep string) func(cfg *Config) (Transformer, bool) {
	return func(cfg *Config) (Transformer, bool) {
		return &numbersCircumfixTransformer{
			cfg: cfg,
			sep: sep,
		}, false
	}
}

type numbersCircumfixTransformer struct {
	cfg *Config
	sep string
}

func (t *numbersCircumfixTransformer) Transform(ctx context.Context, in string) (string, error) {
	if len(in)+len(t.sep) > t.cfg.MaxLen {
		return in, nil
	}

	if v, ok := NumberFromContext(ctx); ok {
		num := strconv.Itoa(v)
		out, ok, err := applyAffix(ctx, t.cfg, circumfix, in, t.sep, num)
		if err != nil || ok {
			return out, err
		}
	}

	stripped, num := StripNumbers(in)
	out, ok, err := applyAffix(ctx, t.cfg, circumfix, stripped, t.sep, num)
	if err != nil || ok {
		return out, nil
	}
	return in, nil
}
