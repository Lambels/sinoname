package sinoname

import (
	"context"
	"errors"
	"math/rand"
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

// errTooLong is a sentinel error. Used by applyAffix to differentiate between false ok value
// returned from source or by the value being too long.
//
// errTooLong should not be returned by transformers, but be used to make further judgement.
var errTooLong = errors.New("value too long")

type affix int

const (
	suffix affix = iota
	prefix
	circumfix
)

func applyAffixFromChunk(ctx context.Context, chunk []int, cfg *Config, where affix, base, sep string) (string, error) {
	n := len(chunk)

	// start with random offset to amplify randomness.
	offset := cfg.RandSrc.Intn(chunks)
	padding := offset * n

	for _, i := range chunk {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		add := cfg.Adjectives[i+padding]
		out, ok, err := applyAffix(ctx, cfg, where, base, sep, add)
		if ok {
			return out, nil
		}
		if err != nil && err != errTooLong {
			return out, err
		}
	}

	rand.Shuffle(len(chunk), func(i, j int) { chunk[i], chunk[j] = chunk[j], chunk[i] })

	for i := 0; i < chunks; i++ {
		// skip random offset (already tried).
		if i == offset {
			continue
		}

		padding = i * n
		for _, j := range chunk {
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			default:
			}

			add := cfg.Adjectives[j+padding]
			out, ok, err := applyAffix(ctx, cfg, where, base, sep, add)
			if ok {
				return out, nil
			}
			if err != nil && err != errTooLong {
				return out, err
			}
		}
		// re shuffle.
		rand.Shuffle(len(chunk), func(i, j int) { chunk[i], chunk[j] = chunk[j], chunk[i] })
	}

	// finish off any values not cought in the chunks.
	remX := len(cfg.Adjectives) % n
	if remX > 0 {
		vals := make([]int, remX)
		for i := range vals {
			vals[i] = i
		}

		rand.Shuffle(len(vals), func(i, j int) { vals[i], vals[j] = vals[j], vals[i] })

		for i := range vals {
			add := cfg.Adjectives[i+chunks*n]
			out, ok, err := applyAffix(ctx, cfg, where, base, sep, add)
			if ok {
				return out, nil
			}
			if err != nil && err != errTooLong {
				return out, err
			}
		}
	}

	return base, nil
}

// applyAffix applies the specified affix to the base. It returns "", false, nil if the lenght is
// to high.
//
// If an error occurs the error is returned, if the value is valid no error is retuned along side true.
func applyAffix(ctx context.Context, cfg *Config, where affix, base, sep, add string) (string, bool, error) {
	switch where {
	case suffix:
		// too long.
		if len(base)+len(sep)+len(add) > cfg.MaxLen {
			return "", false, errTooLong
		}

		out := base + sep + add
		ok, err := cfg.Source.Valid(ctx, out)
		return out, ok, err

	case prefix:
		// too long.
		if len(base)+len(sep)+len(add) > cfg.MaxLen {
			return "", false, errTooLong
		}

		out := add + sep + base
		ok, err := cfg.Source.Valid(ctx, out)
		return out, ok, err

	case circumfix:
		// too long.
		if len(base)+2*len(sep)+2*len(add) > cfg.MaxLen {
			return "", false, errTooLong
		}

		out := add + sep + base + sep + add
		ok, err := cfg.Source.Valid(ctx, out)
		return out, ok, err

	default:
		return "", false, nil
	}
}
