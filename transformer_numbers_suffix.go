package sinoname

import (
	"context"
	"strconv"
)

var NumbersSuffix = func(sep string) func(cfg *Config) (Transformer, bool) {
	return func(cfg *Config) (Transformer, bool) {
		return &numbersSuffixTransformer{
			cfg: cfg,
			sep: sep,
		}, false
	}
}

type numbersSuffixTransformer struct {
	cfg *Config
	sep string
}

func (t *numbersSuffixTransformer) Transform(ctx context.Context, in string) (string, error) {
	if len(in)+len(t.sep) > t.cfg.MaxLen {
		return in, nil
	}

	if v, ok := NumberFromContext(ctx); ok {
		num := strconv.Itoa(v)
		out, ok, err := applyAffix(ctx, t.cfg, suffix, in, t.sep, num)
		if err != nil || ok {
			return out, err
		}
	}

	// len(stripped + num) == len(in)
	stripped, num := StripNumbers(in)
	out, ok, err := applyAffix(ctx, t.cfg, suffix, stripped, t.sep, num)
	if err != nil || ok {
		return out, nil
	}
	return in, nil
}
