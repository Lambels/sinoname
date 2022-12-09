package sinoname

import (
	"context"
	"strconv"
)

// IncrementalPrefix adds an incrementing integer to the end of the string.
// The range of added numbers at the end of the string is [1, n].
//
// Foo1, Foo2, Foo3, Foo4 ... FooN
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

// IncrementalSuffix adds an incrementing integer to the beginning of the string.
// The range of added numbers at the beginning of the string is [1, n].
//
// 1Foo, 2Foo, 3Foo, 4Foo ... NFoo
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

// IncrementalCircumfix adds an incrementing circumfix integer.
// The range of added numbers at the end of the string is [1, n].
//
// 1Foo1, 2Foo2, 3Foo3, 4Foo4 ... NFooN
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

func (t *incrementalTransformer) Transform(ctx context.Context, in MessagePacket) (MessagePacket, error) {
	for i := 1; i <= t.n; i++ {
		add := strconv.Itoa(i)

		out, ok, err := applyAffix(ctx, t.cfg, t.where, in.Message, t.sep, add)
		if ok {
			in.setAndIncrement(out)
			return in, nil
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
