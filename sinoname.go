package sinoname

import (
	"context"
	"errors"
)

// Generator provides extra functionality on top of the layers.
type Generator struct {
	// preventDefault is used to process the data from the layers and omit the default value.
	preventDefault bool

	// maxLen is used both by the generator and layers, it checks that the initial input isnt
	// longer then maxLen.
	maxLen int

	// maxVals buffers the values returned by the layers.
	// the values returned from Generate will either be equal to maxVals or lower.
	maxVals int

	// layers represents the pipeline through which the initial message passes.
	layers Layers
}

// New creates a new generator with the provided config and Layer factories.
func New(conf *Config, layerFacts ...LayerFactory) *Generator {
	g := &Generator{
		maxLen:         conf.MaxLen,
		maxVals:        conf.MaxVals,
		preventDefault: conf.PreventDefault,
	}

	var layers []Layer
	for _, layerFact := range layerFacts {
		if layerFact == nil {
			continue
		}
		l := layerFact(conf)
		layers = append(layers, l)
	}
	g.layers = layers

	return g
}

// Generate passes the in field through the pipeline of transformers. The process can be
// aborted by cancelling the context passed.
func (g *Generator) Generate(ctx context.Context, in string) ([]string, error) {
	if len(in) > g.maxLen {
		return nil, errors.New("sinoname: value is too long")
	}

	inC, clnUp, err := g.layers.Run(ctx, in)
	if err != nil {
		clnUp()
		return nil, err
	}

	var read int
	var vals []string
	for val := range inC {
		if read == g.maxVals {
			break
		}
		if g.preventDefault && val == in {
			continue
		}
		vals = append(vals, val)
		read++
	}

	if err := clnUp(); err != nil {
		return nil, err
	}

	return vals, nil
}
