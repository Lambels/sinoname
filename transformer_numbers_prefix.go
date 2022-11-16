package sinoname

import (
	"context"
	"strconv"
)

var NumbersPrefix = func(sep string) func(cfg *Config) (Transformer, bool) {
	return func(cfg *Config) (Transformer, bool) {
		return &numbersPrefixTransformer{
			cfg: cfg,
			sep: sep,
		}, false
	}
}

type numbersPrefixTransformer struct {
	cfg *Config
	sep string
}

func (t *numbersPrefixTransformer) Transform(ctx context.Context, in string) (string, error) {
	if len(in)+len(t.sep) > t.cfg.MaxLen {
		return in, nil
	}

	if v, ok := NumberFromContext(ctx); ok {
		num := strconv.Itoa(v)
		out, ok, err := applyAffix(ctx, t.cfg, prefix, in, t.sep, num)
		if err != nil || ok {
			return out, err
		}
	}

	// len(stripped + num) == len(in)
	stripped, num := StripNumbers(in)
	out, ok, err := applyAffix(ctx, t.cfg, prefix, stripped, t.sep, num)
	if err != nil || ok {
		return out, nil
	}
	return in, nil
}
