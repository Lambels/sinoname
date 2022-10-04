package sinoname_test

import (
	"context"
	"testing"

	"github.com/Lambels/sinoname"
)

func TestKebabCase(t *testing.T) {
	expected := "kebab-case-test"
	value := "-.kebab -case test"

	tr, _ := sinoname.KebabCase(testConfig)

	v, err := tr.Transform(context.Background(), value)
	if err != nil {
		t.Fatal(err)
	}
	if v != expected {
		t.Fatal("expected:", expected, "but got:", v)
	}
}
