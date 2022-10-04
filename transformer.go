package sinoname

import (
	"context"
)

// Transformer represents a stage of transformation over a message.
//
// The message comes in and comes out modified
type Transformer interface {
	Transform(ctx context.Context, in string) (string, error)
}

// TransformerFactory takes in a config object and returns a transformer.
type TransformerFactory func(cfg *Config) (Transformer, bool)
