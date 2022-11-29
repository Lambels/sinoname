package sinoname

import (
	"context"
	"errors"
	"sync"
)

// Generator provides extra functionality on top of the layers.
type Generator struct {
	// kept a copy for factory functions.
	cfg *Config

	// copies to prevent race conditions.
	preventDuplicates bool
	preventDefault    bool
	maxLen            int
	maxVals           int
	layers            Layers
}

var splitOnDefault = []string{
	".",
	" ",
	"-",
	" ",
	"_",
	" ",
	",",
	" ",
}

// New creates a new generator with the provided config.
func New(conf *Config) *Generator {
	if conf == nil {
		return nil
	}

	if conf.SplitOn == nil {
		conf.SplitOn = splitOnSpecial
	}

	// if adjectives provided, create a pool to share shuffle buffers around all circumfix,
	// suffix or prefix transformer go routines.
	if conf.Adjectives != nil {
		conf.shufflePool = sync.Pool{
			New: func() any {
				chunkSize := len(conf.Adjectives) / chunks
				slc := conf.RandSrc.Perm(chunkSize)
				return slc
			},
		}
	}

	g := &Generator{
		cfg:               conf,
		maxLen:            conf.MaxLen,
		maxVals:           conf.MaxVals,
		preventDuplicates: conf.PreventDuplicates,
		preventDefault:    conf.PreventDefault,
	}

	return g
}

// WithUniformTransformers adds the provided transformers in a uniform layer.
func (g *Generator) WithUniformTransformers(tFact ...TransformerFactory) *Generator {
	uLayer := &UniformTransformerLayer{
		cfg:                  g.cfg,
		transformers:         make([]Transformer, len(tFact)),
		transformerFactories: make([]TransformerFactory, 0),
	}

	for i, f := range tFact {
		t, statefull := f(g.cfg)
		if statefull {
			uLayer.transformerFactories = append(uLayer.transformerFactories, f)
		}
		uLayer.transformers[i] = t
	}
	g.layers = append(g.layers, uLayer)

	return g
}

// WithTransformers adds the provided transformers in a layer (grouped together).
//
// This is the layer configuration which suits most use-cases, you should generally look
// no further.
func (g *Generator) WithTransformers(tFact ...TransformerFactory) *Generator {
	tLayer := &TransformerLayer{
		cfg:                  g.cfg,
		transformers:         make([]Transformer, len(tFact)),
		transformerFactories: make([]TransformerFactory, 0),
	}

	for i, f := range tFact {
		t, statefull := f(g.cfg)
		if statefull {
			tLayer.transformerFactories = append(tLayer.transformerFactories, f)
		}
		tLayer.transformers[i] = t
	}
	g.layers = append(g.layers, tLayer)

	return g
}

// WithLayers adds the provided layers to the generator in order.
func (g *Generator) WithLayers(lFact ...LayerFactory) *Generator {
	for _, f := range lFact {
		l := f(g.cfg)
		g.layers = append(g.layers, l)
	}

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
	readVals := make(map[string]bool)
	readVals[in] = g.preventDefault
L:
	for {
		select {
		// if ctx cancelled no need to call clean up.
		case <-ctx.Done():
			return nil, ctx.Err()

		case val, ok := <-inC:
			if readVals[val] {
				continue
			}
			if g.preventDuplicates {
				readVals[val] = true
			}

			// increment read here so we dont have to wait for next itteration
			// to check if we are at the last value.
			//
			// the next itteration can be slow.
			read++

			if read == g.maxVals || !ok {
				// last value.
				if ok {
					vals = append(vals, val)
				}
				break L
			}
			vals = append(vals, val)
		}
	}

	if err := clnUp(); err != nil {
		// this exception occurs when the maxVals value is reached and
		// there still are live layers.
		if err == context.Canceled {
			return vals, nil
		}

		return nil, err
	}

	return vals, nil
}
