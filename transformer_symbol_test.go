package sinoname_test

import (
	"context"
	"testing"

	"github.com/Lambels/sinoname"
)

// TODO: refactor
func TestSymbol(t *testing.T) {
	t.Run("All_Possible_Values", func(t *testing.T) {
		vals := []string{
			".ABC",
			"A.BC",
			"AB.C",
			"ABC.",
			".A.BC",
			".AB.C",
			".ABC.",
			"A.B.C",
			"A.BC.",
			"AB.C.",
			".A.B.C",
			".A.BC.",
			".AB.C.",
			"A.B.C.",
		}
		src := newCustomSource()
		cfg := &sinoname.Config{
			MaxLen:  testConfig.MaxLen,
			MaxVals: testConfig.MaxVals,
			Source:  src,
		}

		tr := sinoname.SymbolTransformer(".", 3)(cfg)
		for _, vWant := range vals {
			vGot, err := tr.Transform(context.Background(), "ABC")
			if err != nil {
				t.Fatal(err)
			}
			if vGot != vWant {
				t.Fatal("got:", vGot, "but want:", vWant)
			}
			src.addValue(vGot)
		}

		v, err := tr.Transform(context.Background(), "ABC")
		if err != nil {
			t.Fatal(err)
		}
		if v != "ABC" {
			t.Fatal("last iteration wasnt set to initiall value")
		}
	})

	t.Run("Max_Symbols", func(t *testing.T) {
		src := newCustomSource(
			".ABC",
			"A.BC",
			"AB.C",
			"ABC.",
		)
		cfg := &sinoname.Config{
			MaxLen:  testConfig.MaxLen,
			MaxVals: testConfig.MaxVals,
			Source:  src,
		}

		// last itteration should roll to initiall value because no more points can be generated
		// even if possible.
		tr := sinoname.SymbolTransformer(".", 1)(cfg)
		v, err := tr.Transform(context.Background(), "ABC")
		if err != nil {
			t.Fatal(err)
		}
		if v != "ABC" {
			t.Fatal("last iteration wasnt set to initiall value")
		}
	})
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

func newCustomSource(vals ...string) *staticSrc {
	return &staticSrc{
		vals: vals,
	}
}
