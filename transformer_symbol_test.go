package sinoname

import (
	"context"
	"testing"
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
		src := newStaticSource()
		cfg := &Config{
			MaxLen:  testConfig.MaxLen,
			MaxVals: testConfig.MaxVals,
			Source:  src,
		}

		tr, _ := SymbolTransformer('.', 3)(cfg)
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
		src := newStaticSource(
			".ABC",
			"A.BC",
			"AB.C",
			"ABC.",
		)
		cfg := &Config{
			MaxLen:  testConfig.MaxLen,
			MaxVals: testConfig.MaxVals,
			Source:  src,
		}

		// last itteration should roll to initiall value because no more points can be generated
		// even if possible.
		tr, _ := SymbolTransformer('.', 1)(cfg)
		v, err := tr.Transform(context.Background(), "ABC")
		if err != nil {
			t.Fatal(err)
		}
		if v != "ABC" {
			t.Fatal("last iteration wasnt set to initiall value")
		}
	})
}
