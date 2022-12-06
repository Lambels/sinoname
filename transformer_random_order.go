package sinoname

import (
	"context"
	"strings"

	"gonum.org/v1/gonum/stat/combin"
)

//TODO: test if we can start with a random permution and rule it out with an if check.
//TODO: add test.

// ShuffleOrder shuffles the order of the tokens (fields) which make the string up.
// The shuffle isnt random, it takes all the possible permutations for the tokenized input.
//
// Incoming Word: 1980HelloWorld , Separator: _
//
// 1980_Hello_World, 1980_World_Hello, Hello_1980_World, Hello_World_1980, World_1980_Hello, World_Hello_1980
var ShuffleOrder = func(sep string) func(cfg *Config) (Transformer, bool) {
	return func(cfg *Config) (Transformer, bool) {
		return &shuffleOrderTransformer{
			cfg: cfg,
			sep: sep,
		}, false
	}
}

type shuffleOrderTransformer struct {
	cfg *Config
	sep string
}

func (t *shuffleOrderTransformer) Transform(ctx context.Context, in MessagePacket) (MessagePacket, error) {
	split := t.cfg.Tokenize(in.Message)

	// "Lam", "be", "ls" -> Lam_be_ls
	//
	// subtract the extra separator. len("Lam_be_ls_") - len("_") = len("Lam_be_ls")
	if len(split)*len(t.sep)-len(t.sep) > t.cfg.MaxBytes {
		return in, nil
	}

	indexes := make([]int, len(split))
	copyBuf := make([]string, len(split))
	gen := combin.NewPermutationGenerator(len(split), len(split))

	for gen.Next() {
		gen.Permutation(indexes)

		for i, j := range indexes {
			copyBuf[i] = split[j]
		}

		out := strings.Join(copyBuf, t.sep)
		if ok, err := t.cfg.Source.Valid(ctx, out); ok || err != nil {
			in.setAndIncrement(out)
			return in, err
		}
	}

	return in, nil
}
