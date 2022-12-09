package sinoname

import (
	"context"
	"strings"
)

// KebabCase splits on special characters and joins them back with "-".
//
// Foo.Bar_Buz -> Foo-Bar-Buz
var KebabCase = func(cfg *Config) (Transformer, bool) {
	return &kebabCaseTransformer{
		cfg: cfg,
	}, false
}

type kebabCaseTransformer struct {
	cfg *Config
}

func (t *kebabCaseTransformer) Transform(ctx context.Context, in MessagePacket) (MessagePacket, error) {
	split := t.cfg.Tokenize(in.Message)
	out := strings.Join(split, "-")

	if len(out) > t.cfg.MaxBytes {
		return in, nil
	}
	if ok, err := t.cfg.Source.Valid(ctx, out); !ok || err != nil {
		return in, err
	}

	in.setAndIncrement(out)
	return in, nil
}
