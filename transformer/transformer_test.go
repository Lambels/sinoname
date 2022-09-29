package transformer_test

import (
	"context"

	"github.com/Lambels/sinoname/config"
)

var testConfig = &config.Config{
	MaxLen:  100,
	MaxVals: 100,
	Source:  noopSource{},
}

type noopSource struct{}

func (n noopSource) Valid(context.Context, string) (bool, error) {
	return true, nil
}
