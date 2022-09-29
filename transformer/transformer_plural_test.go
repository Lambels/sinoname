package transformer_test

import (
	"context"
	"testing"

	"github.com/Lambels/sinoname/transformer"
)

func TestPlural(t *testing.T) {
	expected := "PluralTests"
	value := "PluralTest"

	tr := transformer.Plural(testConfig)

	v, err := tr.Transform(context.Background(), value)
	if err != nil {
		t.Fatal(err)
	}
	if v != expected {
		t.Fatal("expected:", expected, "but got:", v)
	}
}
