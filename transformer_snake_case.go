package sinoname

var snakeCaseMap map[rune][]rune = map[rune][]rune{
	' ': {'_'},
}

var SnakeCase = func(cfg *Config) Layer {
	layer := &transformerLayer{
		cfg:          cfg,
		transformers: make([]Transformer, 1),
	}
	layer.transformers[0] = &snakeCaseTransformer{
		maxLen: cfg.MaxLen,
		source: cfg.Source,
	}
	return layer
}

type snakeCaseTransformer struct {
	maxLen int
	source Source
}

func (t *snakeCaseTransformer) Transform(in string) (string, error) {
	if len(in) > t.maxLen {
		return in, nil
	}

	split := splitOnSpecial(in)
	out := replaceRunes(split, snakeCaseMap)
	if ok, err := t.source.Valid(out); !ok || err != nil {
		return in, err
	}

	return out, nil
}
