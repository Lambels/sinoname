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
		if len(in)+len(t.sep)+len(num) > t.cfg.MaxLen {
			return in, nil
		}

		out := in + t.sep + num
		ok, err := t.cfg.Source.Valid(ctx, out)
		if err != nil {
			return in, err
		}
		if ok {
			return out, nil
		}
	}

	// len(stripped + num) == len(in)
	stripped, num := StripNumbers(in)
	out := stripped + t.sep + num
	if ok, err := t.cfg.Source.Valid(ctx, out); !ok || err != nil {
		return in, err
	}
	return out, nil
}
