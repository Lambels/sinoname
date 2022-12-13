package sinoname

func newTransformerLayer(tf ...TransformerFactory) *TransformerLayer {
	layer := &TransformerLayer{
		transformers:         make([]Transformer, len(tf)),
		transformerFactories: make([]TransformerFactory, 0),
	}

	for i, f := range tf {
		t, statefull := f(testConfig)
		if statefull {
			layer.transformerFactories = append(layer.transformerFactories, f)
		}
		layer.transformers[i] = t
	}
	return layer
}
