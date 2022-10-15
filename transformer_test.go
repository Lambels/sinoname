package sinoname_test

import (
	"context"
	"testing"

	. "github.com/Lambels/sinoname"
)

var testConfig = &Config{
	MaxLen:  100,
	MaxVals: 100,
	Source:  noopSource{},
	Special: []string{
		".",
		" ",
		"-",
		" ",
		"_",
		" ",
		",",
		" ",
	},
}

type noopSource struct{}

func (n noopSource) Valid(context.Context, string) (bool, error) {
	return true, nil
}

func TestTransformer(t *testing.T) {
	type testCase struct {
		t       TransformerFactory
		in, out string
	}

	var testCases []testCase

	testCases = append(testCases,
		// string cases
		testCase{CamelCase, "-.camel -case test", "CamelCaseTest"},
		testCase{KebabCase, "-.kebab -case test", "kebab-case-test"},
		testCase{SnakeCase, "-.snake -case test", "snake_case_test"},

		testCase{Plural, "plural test", "plural tests"},
		testCase{SymbolTransformer(".", 1), "ABC", ".ABC"},
	)

	for _, tc := range testCases {
		tr, _ := tc.t(testConfig)
		out, err := tr.Transform(context.Background(), tc.in)
		if err != nil {
			t.Fatal(err)
		}

		if out != tc.out {
			t.Fatal("expected:", tc.out, "but got", out)
		}
	}
}
