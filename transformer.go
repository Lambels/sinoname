package sinoname

import (
	"context"
	"errors"
	"time"

	"github.com/Lambels/sinoname/rng"
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
	Transform(ctx context.Context, in MessagePacket) (MessagePacket, error)
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

func applyAffixFromPRNG(ctx context.Context, cfg *Config, gen rng.PRNG, nVals int, where affix, base MessagePacket, sep string, f func(int) string) (MessagePacket, error) {
	offsets := nVals / gen.Range()

	randOffset := cfg.RandSrc.Intn(offsets)
	out, err, done := offsetPRNG(ctx, cfg, randOffset*gen.Range(), gen, where, base, sep, f)
	if done {
		return out, err
	}

	for i := 0; i <= offsets; i++ {
		// skip random offset.
		if i == randOffset {
			continue
		}

		gen.Seed(time.Now().Nanosecond())
		out, err, done := offsetPRNG(ctx, cfg, i*gen.Range(), gen, where, base, sep, f)
		if done {
			return out, err
		}
	}

	// finish off values not cought by range.
	remVals := nVals % gen.Range()
	if remVals == 0 {
		return base, nil
	}

	prngRemVals, clnup := cfg.GetPRNG(remVals)
	defer clnup()

	return applyAffixFromPRNG(ctx, cfg, prngRemVals, remVals, where, base, sep, f)
}

func offsetPRNG(ctx context.Context, cfg *Config, offset int, gen rng.PRNG, where affix, base MessagePacket, sep string, f func(int) string) (MessagePacket, error, bool) {
	for {
		select {
		case <-ctx.Done():
			return base, ctx.Err(), true
		default:
		}

		n, done := gen.Next()
		add := f(n + offset)
		out, ok := applyAffix(cfg, where, base.Message, sep, add)

		if !ok { // if value is too long return if done or continue.
			if done {
				return base, ctx.Err(), false
			}
			continue
		}

		// return if value is valid or error from source.
		if ok, err := cfg.Source.Valid(ctx, out); err != nil || ok {
			base.setAndIncrement(out)
			return base, err, true
		}

		// return if done.
		if done {
			return base, ctx.Err(), false
		}
	}
}

// applyAffix applies the specified affix to the base. It returns "", false, nil if the lenght is
// to high.
//
// If an error occurs the error is returned, if the value is valid no error is retuned along side true.
func applyAffix(cfg *Config, where affix, base, sep, add string) (string, bool) {
	switch where {
	case suffix:
		// too long.
		if len(base)+len(sep)+len(add) > cfg.MaxBytes {
			return "", false
		}

		out := base + sep + add
		return out, true

	case prefix:
		// too long.
		if len(base)+len(sep)+len(add) > cfg.MaxBytes {
			return "", false
		}

		out := add + sep + base
		return out, true

	case circumfix:
		// too long.
		if len(base)+2*len(sep)+2*len(add) > cfg.MaxBytes {
			return "", false
		}

		out := add + sep + base + sep + add
		return out, true

	default:
		return "", false
	}
}
