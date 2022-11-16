package sinoname

import (
	"context"
)

var Suffix = func(sep string) func(cfg *Config) (Transformer, bool) {
	return func(cfg *Config) (Transformer, bool) {
		return &suffixTransformer{
			cfg: cfg,
			sep: sep,
		}, false
	}
}

type suffixTransformer struct {
	cfg *Config
	sep string
}

func (t *suffixTransformer) Transform(ctx context.Context, in string) (string, error) {
	if v, ok := StringFromContext(ctx); ok {
		out, ok, err := applyAffix(ctx, t.cfg, suffix, in, t.sep, v)
		if err != nil || ok {
			return out, err
		}
	}

	// get a single chunk and re use it untill all options exhausted or match found.
	shuffle := t.cfg.getShuffle()
	defer t.cfg.putShuffle(shuffle)

	return applyAffixFromChunk(ctx, shuffle, t.cfg, suffix, in, t.sep)
}
