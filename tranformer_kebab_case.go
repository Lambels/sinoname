package sinoname

var kebabCaseMap map[rune][]rune = map[rune][]rune{
	' ': {'-'},
}

var KebabCase = func(cfg *Config) Transformer {
	return &kebabCaseTransformer{
		maxLen: cfg.MaxLen,
		source: cfg.Source,
	}
}

type kebabCaseTransformer struct {
	maxLen int
	source Source
}

func (t *kebabCaseTransformer) Transform(in string) (string, error) {
	if len(in) > t.maxLen {
		return in, nil
	}

	split := splitOnSpecial(in)
	out := replaceRunes(split, kebabCaseMap)
	if ok, err := t.source.Valid(out); !ok || err != nil {
		return in, err
	}

	return out, nil
}
