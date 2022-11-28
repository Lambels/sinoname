package sinoname

import (
	"context"
	"strings"
	"unicode"
	"unicode/utf8"
)

var AbreviationPrefix = func(all bool) func(cfg *Config) (Transformer, bool) {
	return func(cfg *Config) (Transformer, bool) {
		return &abreviationTransformer{
			cfg:   cfg,
			all:   all,
			where: prefix,
		}, false
	}
}

var AbreviationSuffix = func(all bool) func(cfg *Config) (Transformer, bool) {
	return func(cfg *Config) (Transformer, bool) {
		return &abreviationTransformer{
			cfg:   cfg,
			all:   all,
			where: suffix,
		}, false
	}
}

var AbreviationCircumfix = func(all bool) func(cfg *Config) (Transformer, bool) {
	return func(cfg *Config) (Transformer, bool) {
		return &abreviationTransformer{
			cfg:   cfg,
			all:   all,
			where: circumfix,
		}, false
	}
}

type abreviationTransformer struct {
	cfg   *Config
	all   bool
	where affix
}

func (t *abreviationTransformer) Transform(ctx context.Context, in string) (string, error) {
	if len(in) > t.cfg.MaxLen {
		return in, nil
	}

	split := SplitOnSpecial(in, t.cfg.SplitOn)
	lastX := len(split) - 1
	var out string

	switch t.where {
	case prefix:
		split[0] = ucCapital(split[0])

		if t.all {
			for i := 1; i < lastX; i++ {
				split[i] = ucCapital(split[i])
			}
		}
		out = strings.Join(split, "")

	case suffix:

		split[lastX] = ucCapital(split[lastX])

		if t.all {
			for i := lastX - 1; i > 0; i-- {
				split[i] = ucCapital(split[i])
			}
		}
		out = strings.Join(split, "")

	case circumfix:
		for i := 1; i < lastX; i++ {
			split[i] = ucCapital(split[i])
		}

		if t.all {
			split[0], split[lastX] = ucCapital(split[0]), ucCapital(split[lastX])
		}
		out = strings.Join(split, "")
	}

	ok, err := t.cfg.Source.Valid(ctx, out)
	if ok || err != nil {
		return out, err
	}

	return in, nil
}

func ucCapital(val string) string {
	r, _ := utf8.DecodeRuneInString(val)
	r = unicode.ToUpper(r)

	return string(r)
}
