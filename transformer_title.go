package sinoname

import (
	"context"
	"strings"
	"unicode"
)

// Title capitalizes the first code point in the word.
var Title = func(cfg *Config) (Transformer, bool) {
	return &titleTransformer{
		cfg: cfg,
	}, false
}

type titleTransformer struct {
	cfg *Config
}

func (t *titleTransformer) Transform(ctx context.Context, in MessagePacket) (MessagePacket, error) {
	i := strings.IndexFunc(in.Message, unicode.IsLetter)
	if i == -1 {
		return in, nil
	}

	title := ucCapitalFirst(in.Message[i:])
	out := in.Message[:i] + title

	if ok, err := t.cfg.Source.Valid(ctx, out); !ok || err != nil {
		return in, err
	}

	in.setAndIncrement(out)
	return in, nil
}
