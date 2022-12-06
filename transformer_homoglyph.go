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

// ASCIIHomoglyphLetters maps ascii letters to their homoglyphs.
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

// ASCIIHomoglyphNumbers maps ascii numbers to their homoglyphs.
var ASCIIHomoglyphNumbers ConfidenceMap = ConfidenceMap{
	Map: map[rune][]rune{
		'0': {'O', 'o'},
		'1': {'l', 'I'},
		'3': {'E'},
		'6': {'b'},
	},
	MaxConfidence: 1,
}

// ASCIIHomoglyphSymbols maps ascii symbols to their homoglyphs.
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

var UnicodeHomoglyph map[rune][]rune = map[rune][]rune{
	'a': {'᠑'},
}

var UnicodeHomoglyphSymbols map[rune][]rune = map[rune][]rune{}

// Homoglyph alters the string by replacing runes with their homoglyphs.
//
// The homoglyphs are provided by the confidece maps. The algorithm exhausts each confidence
// map sequentially.
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

func (t *homoglyphTransformer) Transform(ctx context.Context, in MessagePacket) (MessagePacket, error) {
	var maxConfidence int
	for _, v := range t.homoglyphs {
		maxConfidence += v.MaxConfidence
	}

	// CoW implementation.
	for confidence := 0; confidence <= maxConfidence; confidence++ {
		select {
		case <-ctx.Done():
			return MessagePacket{}, ctx.Err()
		default:
		}

		var b strings.Builder
		var next string
		for i, c := range in.Message {
			r := t.getRune(c, confidence)
			if r == c && c != utf8.RuneError {
				continue
			}

			var width int
			if c == utf8.RuneError {
				c, width = utf8.DecodeRuneInString(in.Message[i:])
				// intended RuneError val.
				if width != 1 && r == c {
					continue
				}
			} else {
				width = utf8.RuneLen(c)
			}

			// check if there is space for this rune replacement. (string can grow: ASCII -> UNICODE)
			if n, rWidth := t.cfg.MaxBytes-len(in.Message[i+width:]), utf8.RuneLen(r); n-rWidth < 0 {
				continue
			}

			// changed rune + enough space -> allocate buffer.
			b.Grow(len(in.Message) + utf8.UTFMax)
			b.WriteString(in.Message[:i])
			if r >= 0 {
				b.WriteRune(r)
			}
			// skip current letter.
			next = in.Message[i+width:]
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
			return MessagePacket{}, err
		}
		if ok {
			in.setAndIncrement(out)
			return in, nil
		}

		// write values to buffer whilst always keeping space for at least the bytes remaining
		// in next.
		remainingBytes := t.cfg.MaxBytes - b.Len()
	L:
		for i, c := range next {
			cWidth := utf8.RuneLen(c)
			r := t.getRune(c, confidence)

			var width int
			switch n := remainingBytes - len(next[i+cWidth:]); {
			case n >= utf8.UTFMax: // enough space for any character.
				if r >= 0 {
					if r < utf8.RuneSelf {
						b.WriteByte(byte(r))
						width = 1
					} else {
						width, _ = b.WriteRune(r)
					}
				}

			case n > 0:
				width = utf8.RuneLen(r)

				if width == n { // write and exit (no more space after.).
					if r >= 0 {
						if r < utf8.RuneSelf {
							b.WriteByte(byte(r))
						} else {
							b.WriteRune(r)
						}
					}
					b.WriteString(next[i+cWidth:])
					break L
				} else if width < n { // write and continue (there will still be space left).
					if r >= 0 {
						if r < utf8.RuneSelf {
							b.WriteByte(byte(r))
						} else {
							b.WriteRune(r)
						}
					}
				} else { // write unmodified rune and continue (there is still space for the unmodified rune).
					width, _ = b.WriteRune(c)
				}

			default: // no more space left. write remaining string and break.
				b.WriteString(next[i:])
				break L
			}
			remainingBytes -= width

			// check current variation and see if it is unique.
			out := b.String() + next[i+cWidth:]
			ok, err := t.cfg.Source.Valid(ctx, out)
			if err != nil {
				return MessagePacket{}, err
			}
			if ok {
				in.setAndIncrement(out)
				return in, nil
			}
		}

		// reached when b.Len() == maxBytes.
		out = b.String()
		ok, err = t.cfg.Source.Valid(ctx, out)
		if ok {
			in.setAndIncrement(out)
		}
		return in, err
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
