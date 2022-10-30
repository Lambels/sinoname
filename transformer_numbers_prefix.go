package sinoname

import (
	"context"
	"strconv"
)

var NumbersPrefix = func(separator string) func(cfg *Config) (Transformer, bool) {
	return func(cfg *Config) (Transformer, bool) {
		return &numbersPrefixTransformer{
			sep:    separator,
			maxLen: cfg.MaxLen,
			source: cfg.Source,
		}, false
	}
}

type numbersPrefixTransformer struct {
	sep    string
	maxLen int
	source Source
}

func (t *numbersPrefixTransformer) Transform(ctx context.Context, in string) (string, error) {
	if len(in)+len(t.sep) > t.maxLen {
		return in, nil
	}

	if v, ok := NumberFromContext(ctx); ok {
		num := strconv.Itoa(v)
		if len(in)+len(t.sep)+len(num) > t.maxLen {
			return in, nil
		}

		out := num + t.sep + in
		ok, err := t.source.Valid(ctx, out)
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
	if ok, err := t.source.Valid(ctx, out); !ok || err != nil {
		return in, err
	}
	return out, nil
}
