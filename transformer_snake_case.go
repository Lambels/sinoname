package sinoname

import (
	"context"
	"strings"
)

// SnakeCase splits on special characters and joins them back with "_".
//
// foo.bar_buz -> foo_bar_buz
var SnakeCase = func(cfg *Config) (Transformer, bool) {
	return &snakeCaseTransformer{
		cfg: cfg,
	}, false
}

type snakeCaseTransformer struct {
	cfg *Config
}

func (t *snakeCaseTransformer) Transform(ctx context.Context, in MessagePacket) (MessagePacket, error) {
	split := t.cfg.Tokenize(in.Message)
	out := strings.Join(split, "_")

	if len(out) > t.cfg.MaxBytes {
		return in, nil
	}
	if ok, err := t.cfg.Source.Valid(ctx, out); !ok || err != nil {
		return in, err
	}

	in.setAndIncrement(out)
	return in, nil
}
