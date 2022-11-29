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

func (t *affixShuffleTransformer) Transform(ctx context.Context, in string) (string, error) {
	if v, ok := StringFromContext(ctx); ok {
		out, ok, err := applyAffix(ctx, t.cfg, t.where, in, t.sep, v)
		if err != nil || ok {
			return out, err
		}
	}

	// get a single chunk and re use it untill all options exhausted or match found.
	shuffle := t.cfg.getShuffle()
	defer t.cfg.putShuffle(shuffle)

	return applyAffixFromChunk(ctx, shuffle, t.cfg, t.where, in, t.sep)
}
