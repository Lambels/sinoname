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

var Homoglyph = func(cfg *Config) (Transformer, bool) {
	return &homoglyphTransformer{
		maxLen:     cfg.MaxLen,
		source:     cfg.Source,
		homoglyphs: cfg.SingleHomoglyphTables,
	}, false
}

type homoglyphTransformer struct {
	maxLen     int
	source     Source
	homoglyphs []ConfidenceMap
}

func (t *homoglyphTransformer) Transform(ctx context.Context, in string) (string, error) {
	if len(in)+utf8.UTFMax > t.maxLen {
		return in, nil
	}

	var maxConfidence int
	for _, v := range t.homoglyphs {
		maxConfidence += v.MaxConfidence
	}

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

		// write values to buffer whilst always keeping space for at least
		for {
			c, width := utf8.DecodeRuneInString(next)
			// check if we can accomodate this rune.

			r := t.getRune(c, confidence)
			b.WriteRune()
		}
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
