package transformer

import "github.com/Lambels/sinoname/config"

var Plural = func(cfg *config.Config) Transformer {
	return &pluralTransformer{
		maxLen: cfg.MaxLen,
		source: cfg.Source,
	}
}

type pluralTransformer struct {
	maxLen int
	source config.Source
}

func (t *pluralTransformer) Transform(in string) (string, error) {
	out := in + "s"
	if len(out) > t.maxLen {
		return in, nil
	}

	if ok, err := t.source.Valid(out); !ok || err != nil {
		return in, err
	}

	return out, nil
}
