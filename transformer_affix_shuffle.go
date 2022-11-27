package sinoname

import (
	"context"
)

var Prefix = func(sep string) func(cfg *Config) (Transformer, bool) {
	return func(cfg *Config) (Transformer, bool) {
		return &affixShuffleTransformer{
			where: prefix,
			cfg:   cfg,
			sep:   sep,
		}, false
	}
}

var Suffix = func(sep string) func(cfg *Config) (Transformer, bool) {
	return func(cfg *Config) (Transformer, bool) {
		return &affixShuffleTransformer{
			where: suffix,
			cfg:   cfg,
			sep:   sep,
		}, false
	}
}

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
