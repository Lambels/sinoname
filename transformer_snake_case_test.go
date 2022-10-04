package sinoname_test

import (
	"context"
	"testing"

	"github.com/Lambels/sinoname"
)

func TestSnakeCase(t *testing.T) {
	expected := "snake_case_test"
	value := "-.snake -case test"

	tr, _ := sinoname.SnakeCase(testConfig)

	v, err := tr.Transform(context.Background(), value)
	if err != nil {
		t.Fatal(err)
	}
	if v != expected {
		t.Fatal("expected:", expected, "but got:", v)
	}
}
