package sinoname

import (
	"context"
	"math/rand"
)

var Circumfix = func(sep string) func(cfg *Config) (Transformer, bool) {
	return func(cfg *Config) (Transformer, bool) {
		return &circumfixTransformer{
			cfg: cfg,
			sep: sep,
		}, false
	}
}

type circumfixTransformer struct {
	cfg *Config
	sep string
}

func (t *circumfixTransformer) Transform(ctx context.Context, in string) (string, error) {
	// get a single chunk and re use it untill all options exhausted or match found.
	shuffle := t.cfg.getShuffle()
	defer t.cfg.putShuffle(shuffle)

	n := len(shuffle)
	for i := 0; i < chunks; i++ {
		padding := i * n
		for _, j := range shuffle {
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			default:
			}

			circumfix := t.cfg.Adjectives[j+padding]
			if len(in)+2*len(t.sep)+2*len(circumfix) > t.cfg.MaxLen {
				continue
			}

			out := circumfix + t.sep + in + t.sep + circumfix
			if ok, err := t.cfg.Source.Valid(ctx, out); err != nil || ok {
				return out, err
			}
		}
		// re shuffle.
		rand.Shuffle(len(shuffle), func(i, j int) { shuffle[i], shuffle[j] = shuffle[j], shuffle[i] })
	}

	// finish off any values not cought in the chunks.
	remX := len(t.cfg.Adjectives) % n
	if remX > 0 {
		vals := make([]int, remX)
		for i := range vals {
			vals[i] = i
		}

		rand.Shuffle(len(vals), func(i, j int) { vals[i], vals[j] = vals[j], vals[i] })

		for i := range vals {
			circumfix := t.cfg.Adjectives[i+chunks*n]
			if len(in)+2*len(t.sep)+2*len(circumfix) > t.cfg.MaxLen {
				continue
			}

			out := circumfix + t.sep + in + t.sep + circumfix
			if ok, err := t.cfg.Source.Valid(ctx, out); err != nil || ok {
				return out, err
			}
		}
	}

	return in, nil
}
