package sinoname

import (
	"context"

	"gonum.org/v1/gonum/stat/combin"
)

// SymbolTransformer adds symbol to the incoming word till it fills up any possible
// positions, starting with adding 1 symbol till len(incoming word) - 1  symbols.
//
// Incoming Word: ABC , Symbol: .
//
// Adding 1 Symbol:
// .ABC , A.BC , AB.C , ABC.
//
// Adding 2 Symbols:
// .A.BC , .AB.C , .ABC. , A.B.C , A.BC. , AB.C.
//
// Adding 3 Symbols
// .A.B.C , .A.BC. , .AB.C. , A.B.C.
var SymbolTransformer = func(symbol string, max int) func(cfg *Config) (Transformer, bool) {
	return func(cfg *Config) (Transformer, bool) {
		if max < 0 {
			max = 0
		}

		return &symbolTransformer{
			symbol:     symbol,
			maxLen:     cfg.MaxLen,
			maxSymbols: max,
			source:     cfg.Source,
		}, false
	}
}

type symbolTransformer struct {
	symbol     string
	maxLen     int
	maxSymbols int
	source     Source
}

func (t *symbolTransformer) Transform(ctx context.Context, in string) (string, error) {
	var g *combin.CombinationGenerator
	n := len(in)

	for symbolsToAdd := 1; symbolsToAdd < n+1; symbolsToAdd++ {
		// dont bother to generate if we cant acomodate size after
		// the symbols are added.
		if symbolsToAdd+n > t.maxLen {
			return in, nil
		}
		if symbolsToAdd > t.maxSymbols && t.maxSymbols != 0 {
			return in, nil
		}

		comb := make([]int, symbolsToAdd)
		g = combin.NewCombinationGenerator(n+1, symbolsToAdd)

		for g.Next() {
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			default:
			}

			g.Combination(comb)

			out := applyCombinations(in, comb, t.symbol)
			ok, err := t.source.Valid(ctx, out)
			if err != nil {
				return "", err
			}

			if ok {
				return out, nil
			}
		}
	}

	return in, nil
}

func applyCombinations(in string, comb []int, symbol string) string {
	var offset int
	for _, v := range comb {
		in = in[:v+offset] + symbol + in[v+offset:]
		offset++
	}

	return in
}
