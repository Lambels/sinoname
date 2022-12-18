package sinoname

import (
	"context"
	"strconv"
)

// NumbersPrefix adds a integer to the beginning of the string.
// It obtains the integer by:
//  1. From the context via NumberFromContext .
//  2. Collects all the numbers from the string.
//
// Foo1 Bar2 Buz3 -> 123Foo Bar Buz .
var NumbersPrefix = func(sep string) func(cfg *Config) (Transformer, bool) {
	return func(cfg *Config) (Transformer, bool) {
		return &numbersTransformer{
			where: prefix,
			cfg:   cfg,
			sep:   sep,
		}, false
	}
}

// NumbersSuffix adds a integer to the end of the string.
// It obtains the integer by:
//  1. From the context via NumberFromContext .
//  2. Collects all the numbers from the string.
//
// Foo1 Bar2 Buz3 -> Foo Bar Buz123
var NumbersSuffix = func(sep string) func(cfg *Config) (Transformer, bool) {
	return func(cfg *Config) (Transformer, bool) {
		return &numbersTransformer{
			where: suffix,
			cfg:   cfg,
			sep:   sep,
		}, false
	}
}

// NumbersCircumfix adds a integer to the end and beinning of the string.
// It obtains the integer by:
//  1. From the context via NumberFromContext .
//  2. Collects all the numbers from the string.
//
// Foo1 Bar2 Buz3 -> 123Foo Bar Buz123
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

func (t *numbersTransformer) Transform(ctx context.Context, in MessagePacket) (MessagePacket, error) {
	if len(in.Message)+len(t.sep) > t.cfg.MaxBytes {
		return in, nil
	}

	if v, ok := NumberFromContext(ctx); ok {
		num := strconv.Itoa(v)
		out, ok := applyAffix(ctx, t.cfg, t.where, in.Message, t.sep, num)
		if ok {
			unique, err := t.cfg.Source.Valid(ctx, out)
			if err != nil || unique {
				in.setAndIncrement(out)
				return in, err
			}
		}
	}

	stripped, num := t.cfg.StripNumbers(in.Message)
	// no point in checking the ok value since len(stripped) + len(num) == len(in.Message)
	out, _ := applyAffix(ctx, t.cfg, t.where, stripped, t.sep, num)
	ok, err := t.cfg.Source.Valid(ctx, out)
	if ok {
		in.setAndIncrement(out)
	}

	return in, err
}
