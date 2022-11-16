package sinoname

import (
	"context"
	"testing"
)

func TestTransformer(t *testing.T) {
	type testCase struct {
		ctx     context.Context
		t       TransformerFactory
		in, out string
	}

	var testCases []testCase

	testCases = append(testCases,
		testCase{t: CamelCase, in: "-.camel -case test", out: "CamelCaseTest"},
		testCase{t: KebabCase, in: "-.kebab -case test", out: "kebab-case-test"},
		testCase{t: SnakeCase, in: "-.snake -case test", out: "snake_case_test"},

		testCase{t: Plural, in: "plural test", out: "plural tests"},
		testCase{t: SymbolTransformer('.', 1), in: "ABC", out: ".ABC"},
		testCase{ContextWithNumber(context.Background(), 100), NumbersPrefix("_"), "Patrick1234", "100_Patrick1234"},
		testCase{t: NumbersPrefix("-"), in: "Patrick1234", out: "1234-Patrick"},
		testCase{ContextWithNumber(context.Background(), 100), NumbersSuffix("_"), "1234Patrick", "1234Patrick_100"},
		testCase{t: NumbersSuffix("-"), in: "Patrick1234", out: "Patrick-1234"},
		testCase{t: Homoglyph(ASCIIHomoglyphLetters), in: "bee", out: "6ee"},
	)

	// evaluate test cases.
	for _, tc := range testCases {
		if tc.ctx == nil {
			tc.ctx = context.Background()
		}

		tr, _ := tc.t(testConfig)

		out, err := tr.Transform(tc.ctx, tc.in)
		if err != nil {
			t.Fatal(err)
		}

		if out != tc.out {
			t.Fatal("expected:", tc.out, "but got", out)
		}
	}
}
