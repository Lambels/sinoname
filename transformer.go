package sinoname

import (
	"context"
	"errors"
)

// ErrSkip should be used by transformers to skip the output and not pass it
// further down the pipeline.
var ErrSkip error = errors.New("skip output")

// Transformer represents a stage of transformation over a message.
//
// The message comes in and comes out modified.
//
// A trasnformer should handle context cancellations if possible and return any
// errors from the source.
type Transformer interface {
	Transform(ctx context.Context, in string) (string, error)
}

// TransformerFactory takes in a config object and returns a transformer and a
// state indicator.
//
// If the state indicator has true boolean value then the trasnformer layer using it is
// going to create a new Transformer per each (sinoname.Layer).PumpOut() call.
//
// For most transformers no state value is required since transformers by nature should be
// simple and closest to a pure function. Although the option for a statefull transformer
// is provided and suported by all layers.
type TransformerFactory func(cfg *Config) (Transformer, bool)
