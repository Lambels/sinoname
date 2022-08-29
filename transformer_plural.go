package sinoname

var Plural = func(cfg *Config) Layer {
	layer := &transformerLayer{
		cfg:          cfg,
		transformers: make([]Transformer, 1),
	}
	layer.transformers[0] = &pluralTransformer{
		maxLen: cfg.MaxLen,
		source: cfg.Source,
	}
	return layer
}

type pluralTransformer struct {
	maxLen int
	source Source
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
