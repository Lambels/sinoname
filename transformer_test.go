package sinoname_test

import (
	"context"

	"github.com/Lambels/sinoname"
)

var testConfig = &sinoname.Config{
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
