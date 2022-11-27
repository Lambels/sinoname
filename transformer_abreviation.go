package sinoname

import (
	"context"
	"strings"
	"unicode"
)

//TODO: add all option -> FOO BAR BUZ -> FBBUZ // FBARBUZ
//TODO: all option for circumfix -> FOO BAR BUZ -> FBB

var AbreviationPrefix = func(cfg *Config) (Transformer, bool) {
	return &abreviationTransformer{
		cfg:   cfg,
		where: prefix,
	}, false
}

var AbreviationSuffix = func(cfg *Config) (Transformer, bool) {
	return &abreviationTransformer{
		cfg:   cfg,
		where: suffix,
	}, false
}

var AbreviationCircumfix = func(cfg *Config) (Transformer, bool) {
	return &abreviationTransformer{
		cfg:   cfg,
		where: circumfix,
	}, false
}

type abreviationTransformer struct {
	cfg   *Config
	where affix
}

func (t *abreviationTransformer) Transform(ctx context.Context, in string) (string, error) {
	if len(in) > t.cfg.MaxLen {
		return in, nil
	}

	split := SplitOnSpecial(in, t.cfg.SplitOn)
	var out string

	switch t.where {
	case prefix:
		split[0] = string(unicode.ToUpper(rune(split[0][0])))
		out = strings.Join(split, "")

	case suffix:
		lastX := len(split) - 1

		split[lastX] = string(unicode.ToUpper(rune(
			split[lastX][len(split[lastX])-1],
		)))
		out = strings.Join(split, "")

	case circumfix:
		for i := 1; i < len(split)-1; i++ {
			split[i] = string(unicode.ToUpper(rune(split[i][0])))
		}
		out = strings.Join(split, "")
	}

	if len(out) > t.cfg.MaxLen {
		return in, nil
	}

	ok, err := t.cfg.Source.Valid(ctx, out)
	if ok || err != nil {
		return out, err
	}

	return in, nil
}
