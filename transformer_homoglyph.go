package sinoname

import (
	"context"
	"strings"
	"unicode/utf8"
)

type ConfidenceMap struct {
	Map           map[rune][]rune
	MaxConfidence int
}

var ASCIIHomoglyphLetters ConfidenceMap = ConfidenceMap{
	Map: map[rune][]rune{
		'b': {'6'},
		'c': {'C'},
		'e': {'3'},
		'i': {'1', 'l'},
		'l': {'I', '1'},
		'o': {'O', '0'},
		'q': {'g'},
		's': {'S', '5', 'z'},
		'u': {'v', 'U'},
		'v': {'u', 'V'},
		'w': {'W'},
		'z': {'s'},

		'A': {'4'},
		'B': {'8'},
		'C': {'c'},
		'E': {'3'},
		'I': {'l', '1'},
		'S': {'s', '5', '2'},
		'U': {'V', 'u'},
		'V': {'U', 'v'},
		'O': {'0', 'o', 'Q'},
		'Z': {'2'},
	},
	MaxConfidence: 2,
}

var ASCIIHomoglyphNumbers ConfidenceMap = ConfidenceMap{
	Map: map[rune][]rune{
		'0': {'O', 'o'},
		'1': {'l', 'I'},
		'3': {'E'},
		'6': {'b'},
	},
	MaxConfidence: 1,
}

var ASCIIHomoglyphSymbols ConfidenceMap = ConfidenceMap{
	Map: map[rune][]rune{
		'a': {'&', '@'},
		's': {'$'},
		'l': {'|'},

		'S': {'$'},

		'$': {'S', 's'},
	},
	MaxConfidence: 1,
}

var UnicodeHomoglyph map[rune][]rune = map[rune][]rune{}

var UnicodeHomoglyphSymbols map[rune][]rune = map[rune][]rune{}

var Homoglyph = func(homoglyphs ...ConfidenceMap) func(cfg *Config) (Transformer, bool) {
	return func(cfg *Config) (Transformer, bool) {
		return &homoglyphTransformer{
			cfg:        cfg,
			homoglyphs: homoglyphs,
		}, false
	}
}

type homoglyphTransformer struct {
	cfg        *Config
	homoglyphs []ConfidenceMap
}

func (t *homoglyphTransformer) Transform(ctx context.Context, in string) (string, error) {
	if len(in)+utf8.UTFMax > t.cfg.MaxLen {
		return in, nil
	}

	var maxConfidence int
	for _, v := range t.homoglyphs {
		maxConfidence += v.MaxConfidence
	}

	// CoW implementation.
	for confidence := 0; confidence <= maxConfidence; confidence++ {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		var b strings.Builder
		var next string
		for i, c := range in {
			r := t.getRune(c, confidence)
			if r == c && c != utf8.RuneError {
				continue
			}

			var width int
			if c == utf8.RuneError {
				c, width = utf8.DecodeRuneInString(in[i:])
				// intended RuneError val.
				if width != 1 && r == c {
					continue
				}
			} else {
				width = utf8.RuneLen(c)
			}

			// changed rune, allocate buffer.
			b.Grow(len(in) + utf8.UTFMax)
			b.WriteString(in[:i])

			if r >= 0 {
				b.WriteRune(r)
			}

			// skip current letter.
			next = in[i+width:]
			break
		}

		// capacity is 0. no buffer was initialized therefor value unchanged.
		if b.Cap() == 0 {
			return in, nil
		}

		// check if string is valid after first modification.
		out := b.String() + next
		ok, err := t.cfg.Source.Valid(ctx, out)
		if err != nil {
			return "", err
		}
		if ok {
			return out, nil
		}

		// write values to buffer whilst always keeping space for at least the bytes remaining
		// in next.
		remainingBytes := t.cfg.MaxLen - b.Len()
		for i, c := range next {
			r := t.getRune(c, confidence)

			var width int
			if n := remainingBytes - len(next[i:]); n < utf8.UTFMax {
				width = utf8.RuneLen(r)
				if width > n {
					b.WriteString(next[i:])
					break
				}
				if r >= 0 {
					if r < utf8.RuneSelf {
						b.WriteByte(byte(r))
					} else {
						// r is not a ASCII rune.
						b.WriteRune(r)
					}
				}
			} else if r >= 0 {
				if r < utf8.RuneSelf {
					b.WriteByte(byte(r))
					width = 1
				} else {
					// r is not a ASCII rune.
					width, _ = b.WriteRune(r)
				}
			}

			remainingBytes -= width
			out := b.String() + next[i+width:]
			ok, err := t.cfg.Source.Valid(ctx, out)
			if err != nil {
				return "", err
			}
			if ok {
				return out, nil
			}
		}

		out = b.String()
		_, err = t.cfg.Source.Valid(ctx, out)
		return out, err
	}

	return in, nil
}

// replace replaces the rune at index i with a mapped runed at a given confidence.
// returns how many bytes were added to the string.
func (t *homoglyphTransformer) getRune(c rune, confidence int) rune {
	for _, mapping := range t.homoglyphs {
		// there is no possible way that this confidence is supported in
		// this mapping.
		if mapping.MaxConfidence < confidence {
			confidence -= mapping.MaxConfidence
			continue
		}

		replaceRunes, ok := mapping.Map[c]
		// no possible confidence level, return.
		if !ok || len(replaceRunes) < confidence {
			return c
		}

		return replaceRunes[confidence]
	}

	return c
}
