package sinoname_test

import (
	"context"
	"testing"

	"github.com/Lambels/sinoname"
)

func TestCamelCase(t *testing.T) {
	expected := "CamelCaseTest"
	value := "-.camel -case test"

	tr := sinoname.CamelCase(testConfig)

	v, err := tr.Transform(context.Background(), value)
	if err != nil {
		t.Fatal(err)
	}
	if v != expected {
		t.Fatal("expected:", expected, "but got:", v)
	}
}
