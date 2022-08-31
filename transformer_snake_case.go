package sinoname

var snakeCaseMap map[rune][]rune = map[rune][]rune{
	' ': {'_'},
}

var SnakeCase = func(cfg *Config) Transformer {
	return &snakeCaseTransformer{
		maxLen: cfg.MaxLen,
		source: cfg.Source,
	}
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
