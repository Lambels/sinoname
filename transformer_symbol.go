package sinoname

import (
	"context"
	"strings"
	"unicode/utf8"

	"gonum.org/v1/gonum/stat/combin"
)

//TODO: test if for each symbol to add we can generate a random permutation and then
//TODO: check for it (and skip it) in the sequential passings of the permuations to randomise things up.

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
var SymbolTransformer = func(symbol rune, max int) func(cfg *Config) (Transformer, bool) {
	return func(cfg *Config) (Transformer, bool) {
		if max < 0 {
			max = 0
		}

		return &symbolTransformer{
			cfg:        cfg,
			symbol:     symbol,
			maxSymbols: max,
		}, false
	}
}

type symbolTransformer struct {
	cfg        *Config
	symbol     rune
	maxSymbols int
}

func (t *symbolTransformer) Transform(ctx context.Context, in MessagePacket) (MessagePacket, error) {
	var g *combin.CombinationGenerator
	n := len(in.Message)
	nr := utf8.RuneLen(t.symbol)

	for symbolsToAdd := 1; symbolsToAdd < n+1; symbolsToAdd++ {
		// dont bother to generate and allocate buffer if we cant acomodate size after
		// the symbols are added.
		if n+symbolsToAdd*nr > t.cfg.MaxBytes {
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
				return MessagePacket{}, ctx.Err()
			default:
			}

			var b strings.Builder
			b.Grow(n + symbolsToAdd*nr)

			g.Combination(comb)
			var prevJ int
			for i, j := range comb {
				b.WriteString(in.Message[prevJ:j])
				b.WriteRune(t.symbol)
				prevJ = j

				// last itteration, write remaining string.
				if i == len(comb)-1 {
					b.WriteString(in.Message[j:])
				}
			}

			out := b.String()
			ok, err := t.cfg.Source.Valid(ctx, out)
			if err != nil {
				return MessagePacket{}, err
			}

			if ok {
				in.setAndIncrement(out)
				return in, nil
			}
		}
	}

	return in, nil
}
