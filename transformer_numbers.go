package sinoname

import (
	"context"
	"strconv"
)

var NumbersPrefix = func(sep string) func(cfg *Config) (Transformer, bool) {
	return func(cfg *Config) (Transformer, bool) {
		return &numbersTransformer{
			where: prefix,
			cfg:   cfg,
			sep:   sep,
		}, false
	}
}

var NumbersSuffix = func(sep string) func(cfg *Config) (Transformer, bool) {
	return func(cfg *Config) (Transformer, bool) {
		return &numbersTransformer{
			where: suffix,
			cfg:   cfg,
			sep:   sep,
		}, false
	}
}

var NumbersCircumfix = func(sep string) func(cfg *Config) (Transformer, bool) {
	return func(cfg *Config) (Transformer, bool) {
		return &numbersTransformer{
			where: circumfix,
			cfg:   cfg,
			sep:   sep,
		}, false
	}
}

type numbersTransformer struct {
	where affix
	cfg   *Config
	sep   string
}

func (t *numbersTransformer) Transform(ctx context.Context, in string) (string, error) {
	if len(in)+len(t.sep) > t.cfg.MaxLen {
		return in, nil
	}

	if v, ok := NumberFromContext(ctx); ok {
		num := strconv.Itoa(v)
		out, ok, err := applyAffix(ctx, t.cfg, t.where, in, t.sep, num)
		if ok {
			return out, nil
		}

		if err != nil && err != errTooLong {
			return out, err
		}
	}

	// len(stripped + num) == len(in)
	stripped, num := StripNumbers(in)
	out, ok, err := applyAffix(ctx, t.cfg, t.where, stripped, t.sep, num)
	if ok {
		return out, nil
	}

	// errTooLong voided.
	if err == errTooLong {
		err = nil
	}

	return in, err
}
