package sinoname

import (
	"context"
	"strings"
	"unicode"
	"unicode/utf8"
)

// AbreviationPrefix abreviates words starting from the left going to the right.
//
// An extra parameter is accepted: all , when true all the words from the left are abreviated
// but the last one.
//
// Examples:
//
// Foo-Bar Buz -> FBarBuz (all = false)
//
// Foo-Bar Buz -> FBBuz (all = true)
var AbreviationPrefix = func(sep string, all bool) func(cfg *Config) (Transformer, bool) {
	return func(cfg *Config) (Transformer, bool) {
		return &abreviationTransformer{
			cfg:   cfg,
			all:   all,
			sep:   sep,
			where: prefix,
		}, false
	}
}

// AbreviationSuffix abreviates words starting from the right going to the left.
//
// An extra parameter is accepted: all , when true all the words from the right are abreviated
// but the first one.
//
// Examples:
//
// Foo-Bar Buz -> FooBarB (all = false)
//
// Foo-Bar Buz -> FooBB (all = true)
var AbreviationSuffix = func(sep string, all bool) func(cfg *Config) (Transformer, bool) {
	return func(cfg *Config) (Transformer, bool) {
		return &abreviationTransformer{
			cfg:   cfg,
			all:   all,
			sep:   sep,
			where: suffix,
		}, false
	}
}

// AbreviationCircumfix abreviates the words in the middle of the string.
//
// An extra parameter is accepted: all , when true all the words are abreviated
//
// Examples:
//
// Foo-Bar Fuz Buz -> FooBFBuz (all = false)
//
// Foo-Bar Fuz Buz -> FBFB (all = true)
var AbreviationCircumfix = func(sep string, all bool) func(cfg *Config) (Transformer, bool) {
	return func(cfg *Config) (Transformer, bool) {
		return &abreviationTransformer{
			cfg:   cfg,
			all:   all,
			sep:   sep,
			where: circumfix,
		}, false
	}
}

type abreviationTransformer struct {
	cfg   *Config
	all   bool
	sep   string
	where affix
}

func (t *abreviationTransformer) Transform(ctx context.Context, in MessagePacket) (MessagePacket, error) {
	if len(in.Message) > t.cfg.MaxBytes {
		return in, nil
	}

	split := t.cfg.Tokenize(in.Message)
	lastX := len(split) - 1

	switch t.where {
	case prefix:
		split[0] = ucCapital(split[0])

		if t.all {
			for i := 1; i < lastX; i++ {
				split[i] = ucCapital(split[i])
			}
		}

	case suffix:
		split[lastX] = ucCapital(split[lastX])

		if t.all {
			for i := lastX - 1; i > 0; i-- {
				split[i] = ucCapital(split[i])
			}
		}

	case circumfix:
		for i := 1; i < lastX; i++ {
			split[i] = ucCapital(split[i])
		}

		if t.all {
			split[0], split[lastX] = ucCapital(split[0]), ucCapital(split[lastX])
		}
	}

	out := strings.Join(split, t.sep)

	if len(out) > t.cfg.MaxBytes {
		return in, nil
	}

	ok, err := t.cfg.Source.Valid(ctx, out)
	if ok || err != nil {
		in.setAndIncrement(out)
		return in, err
	}

	return in, nil
}

func ucCapital(val string) string {
	r, _ := utf8.DecodeRuneInString(val)
	r = unicode.ToUpper(r)

	return string(r)
}
