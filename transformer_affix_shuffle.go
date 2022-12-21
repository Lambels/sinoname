package sinoname

import (
	"context"
)

// Prefix adds a prefix to the string.
//
// The prefix is obtained:
//  1. From the context via the StringFromContext method.
//  2. At random from the adjectives array provided in the config object.
var Prefix = func(sep string) func(cfg *Config) (Transformer, bool) {
	return func(cfg *Config) (Transformer, bool) {
		return &affixShuffleTransformer{
			where: prefix,
			cfg:   cfg,
			sep:   sep,
		}, false
	}
}

// Suffix adds a suffix to the string.
//
// The suffix is obtained:
//  1. From the context via the StringFromContext method.
//  2. At random from the adjectives array provided in the config object.
var Suffix = func(sep string) func(cfg *Config) (Transformer, bool) {
	return func(cfg *Config) (Transformer, bool) {
		return &affixShuffleTransformer{
			where: suffix,
			cfg:   cfg,
			sep:   sep,
		}, false
	}
}

// Circumfix adds a circumfix to the string.
//
// The circumfix is obtained:
//  1. From the context via the StringFromContext method.
//  2. At random from the adjectives array provided in the config object.
var Circumfix = func(sep string) func(cfg *Config) (Transformer, bool) {
	return func(cfg *Config) (Transformer, bool) {
		return &affixShuffleTransformer{
			where: circumfix,
			cfg:   cfg,
			sep:   sep,
		}, false
	}
}

type affixShuffleTransformer struct {
	cfg   *Config
	where affix
	sep   string
}

func (t *affixShuffleTransformer) Transform(ctx context.Context, in MessagePacket) (MessagePacket, error) {
	if v, ok := StringFromContext(ctx); ok {
		out, ok := applyAffix(t.cfg, t.where, in.Message, t.sep, v)

		if ok {
			unique, err := t.cfg.Source.Valid(ctx, out)
			if err != nil || unique {
				in.setAndIncrement(out)
				return in, err
			}
		}
	}

	prng, clnup := t.cfg.GetPRNG(len(t.cfg.Adjectives))
	defer clnup()

	f := func(i int) string {
		return t.cfg.Adjectives[i]
	}
	return applyAffixFromPRNG(ctx, t.cfg, prng, len(t.cfg.Adjectives), t.where, in, t.sep, f)
}
