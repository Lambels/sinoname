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
		if len(in)+len(t.sep)+len(num) > t.cfg.MaxLen {
			return in, nil
		}

		out := num + t.sep + in
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
	out := num + t.sep + stripped
	if ok, err := t.cfg.Source.Valid(ctx, out); !ok || err != nil {
		return in, err
	}
	return out, nil
}
