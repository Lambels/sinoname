package sinoname

import (
	"context"
	"strconv"
)

var IncrementalPrefix = func(n int, sep string) func(cfg *Config) (Transformer, bool) {
	return func(cfg *Config) (Transformer, bool) {
		return &incrementalTransformer{
			cfg:   cfg,
			where: prefix,
			n:     n,
			sep:   sep,
		}, false
	}
}

var IncrementalSuffix = func(n int, sep string) func(cfg *Config) (Transformer, bool) {
	return func(cfg *Config) (Transformer, bool) {
		return &incrementalTransformer{
			cfg:   cfg,
			where: suffix,
			n:     n,
			sep:   sep,
		}, false
	}
}

var IncrementalCircumfix = func(n int, sep string) func(cfg *Config) (Transformer, bool) {
	return func(cfg *Config) (Transformer, bool) {
		return &incrementalTransformer{
			cfg:   cfg,
			where: circumfix,
			n:     n,
			sep:   sep,
		}, false
	}
}

type incrementalTransformer struct {
	cfg   *Config
	where affix
	n     int
	sep   string
}

func (t *incrementalTransformer) Transform(ctx context.Context, in string) (string, error) {
	if len(in) > t.cfg.MaxLen {
		return in, nil
	}

	for i := 1; i <= t.n; i++ {
		add := strconv.Itoa(i)

		out, ok, err := applyAffix(ctx, t.cfg, t.where, in, t.sep, add)
		if ok {
			return out, nil
		}

		if err != nil {
			// retrun early even if value too long.
			// values only continue growing, no point in continuing.
			if err == errTooLong {
				err = nil
			}

			return in, err
		}
	}

	return in, nil
}
