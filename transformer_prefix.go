package sinoname

import (
	"context"
)

var Prefix = func(sep string) func(cfg *Config) (Transformer, bool) {
	return func(cfg *Config) (Transformer, bool) {
		return &prefixTransformer{
			cfg: cfg,
			sep: sep,
		}, false
	}
}

type prefixTransformer struct {
	cfg *Config
	sep string
}

func (t *prefixTransformer) Transform(ctx context.Context, in string) (string, error) {
	if v, ok := StringFromContext(ctx); ok {
		out, ok, err := applyAffix(ctx, t.cfg, prefix, in, t.sep, v)
		if err != nil || ok {
			return out, err
		}
	}

	// get a single chunk and re use it untill all options exhausted or match found.
	shuffle := t.cfg.getShuffle()
	defer t.cfg.putShuffle(shuffle)

	return applyAffixFromChunk(ctx, shuffle, t.cfg, prefix, in, t.sep)
}
