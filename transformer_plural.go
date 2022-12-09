package sinoname

import (
	"context"
)

// Plural adds an s at the end of the string.
var Plural = func(cfg *Config) (Transformer, bool) {
	return &pluralTransformer{
		cfg: cfg,
	}, false
}

type pluralTransformer struct {
	cfg *Config
}

func (t *pluralTransformer) Transform(ctx context.Context, in MessagePacket) (MessagePacket, error) {
	if len(in.Message)+1 > t.cfg.MaxBytes {
		return in, nil
	}
	out := in.Message + "s"

	if ok, err := t.cfg.Source.Valid(ctx, out); !ok || err != nil {
		return in, err
	}

	in.setAndIncrement(out)
	return in, nil
}
