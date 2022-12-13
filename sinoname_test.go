package sinoname

import (
	"context"
)

var testConfig = &Config{
	MaxBytes:     100,
	MaxVals:      100,
	Source:       noopSource{true},
	Tokenize:     tokenizeDefault,
	StripNumbers: stripNumbersASCII,
}

type noopSource struct{ b bool }

func (n noopSource) Valid(context.Context, string) (bool, error) {
	return n.b, nil
}

type staticSrc struct {
	vals []string
}

func (s *staticSrc) addValue(v string) {
	s.vals = append(s.vals, v)
}

func (s *staticSrc) Valid(ctx context.Context, in string) (bool, error) {
	for _, v := range s.vals {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		default:
		}

		if v == in {
			return false, nil
		}
	}

	return true, nil
}

func newStaticSource(vals ...string) *staticSrc {
	return &staticSrc{
		vals: vals,
	}
}
