package sinoname

import (
	"context"
	"errors"

	"github.com/Lambels/sinoname/config"
	"github.com/Lambels/sinoname/layer"
	"github.com/Lambels/sinoname/transformer"
)

// Generator provides extra functionality on top of the layers.
type Generator struct {
	// kept a copy for factory functions.
	cfg *config.Config

	// preventDefault is used to process the data from the layers and omit the default value.
	preventDefault bool

	// maxLen is used both by the generator and layers, it checks that the initial input isnt
	// longer then maxLen.
	maxLen int

	// maxVals buffers the values returned by the layers.
	// the values returned from Generate will either be equal to maxVals or lower.
	maxVals int

	// layers represents the pipeline through which the initial message passes.
	layers layer.Layers
}

// New creates a new generator with the provided config and Layer factories.
func New(conf *config.Config) *Generator {
	g := &Generator{
		cfg:            conf,
		maxLen:         conf.MaxLen,
		maxVals:        conf.MaxVals,
		preventDefault: conf.PreventDefault,
	}

	return g
}

func (g *Generator) WithUniformTransformers(tFact ...layer.TransformerFactory) *Generator {
	uLayer := &layer.UniformTransformerLayer{
		Transformers: make([]transformer.Transformer, len(tFact)),
	}

	for i, f := range tFact {
		t := f(g.cfg)
		uLayer.Transformers[i] = t
	}
	g.layers = append(g.layers, uLayer)

	return g
}

func (g *Generator) WithTransformers(tFact ...layer.TransformerFactory) *Generator {
	tLayer := &layer.TransformerLayer{
		Transformers: make([]transformer.Transformer, len(tFact)),
	}

	for i, f := range tFact {
		t := f(g.cfg)
		tLayer.Transformers[i] = t
	}
	g.layers = append(g.layers, tLayer)

	return g
}

// WithProxys creates a new proxy layer and adds it to the generator which fans in all messages
// from the parent layer and runs the proxy functions on each message.
//
// For a message to pass through the proxy layer it must pass through all the proxy functions
// without returning any error, else the message is just consumed from the upstream layer
// and not sent further.
func (g *Generator) WithProxys(pFact ...layer.ProxyFactory) *Generator {
	pLayer := &layer.ProxyLayer{
		Proxys: make([]layer.ProxyFunc, len(pFact)),
	}

	for i, f := range pFact {
		t := f(g.cfg)
		pLayer.Proxys[i] = t
	}
	g.layers = append(g.layers, pLayer)
	return g
}

func (g *Generator) WithLayers(lFact ...layer.LayerFactory) *Generator {
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
			vals = append(vals, val)

			if read == g.maxVals || !ok {
				break L
			}
			if g.preventDefault && val == in {
				continue
			}
		}
	}

	if err := clnUp(); err != nil {
		return nil, err
	}

	return vals, nil
}
