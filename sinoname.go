package sinoname

import (
	"context"
	"errors"
)

// Generator provides extra functionality on top of the layers.
type Generator struct {
	// kept a copy for factory functions.
	cfg *Config

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

// New creates a new generator with the provided config and Layer factories.
func New(conf *Config) *Generator {
	if conf == nil {
		return nil
	}

	if conf.SplitOn == nil {
		conf.SplitOn = splitOnDefault
	} else {
		oldNew := make([]string, len(conf.SplitOn)*2)
		for i, v := range conf.SplitOn {
			oldNew[i*2] = v
			oldNew[i*2+1] = " "
		}
		conf.SplitOn = oldNew
	}

	g := &Generator{
		cfg:            conf,
		maxLen:         conf.MaxLen,
		maxVals:        conf.MaxVals,
		preventDefault: conf.PreventDefault,
	}

	return g
}

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
L:
	for {
		select {
		// if ctx cancelled no need to call clean up.
		case <-ctx.Done():
			return nil, ctx.Err()

		case val, ok := <-inC:
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
			if g.preventDefault && val == in {
				continue
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
