package sinoname

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"
)

type ConfidenceMap struct {
	Map           map[rune][][]rune
	MaxConfidence int
}

// ASCIIHomoglyphLetters maps ascii letters to their homoglyphs.
var ASCIIHomoglyphLetters ConfidenceMap = ConfidenceMap{
	Map: map[rune][][]rune{
		'b': {{'6'}},
		'c': {{'C'}},
		'e': {{'3'}},
		'i': {{'1'}, {'l'}},
		'l': {{'I'}, {'1'}},
		'o': {{'O'}, {'0'}},
		'q': {{'g'}},
		's': {{'S'}, {'5'}, {'z'}},
		'u': {{'v'}, {'U'}},
		'v': {{'u'}, {'V'}},
		'w': {{'W'}},
		'z': {{'s'}},

		'A': {{'4'}},
		'B': {{'8'}},
		'C': {{'c'}},
		'E': {{'3'}},
		'I': {{'l'}, {'1'}},
		'S': {{'s'}, {'5'}, {'2'}},
		'U': {{'V'}, {'u'}},
		'V': {{'U'}, {'v'}},
		'O': {{'0'}, {'o'}, {'Q'}},
		'Z': {{'2'}},
	},
	MaxConfidence: 2,
}

// ASCIIHomoglyphNumbers maps ascii numbers to their homoglyphs.
var ASCIIHomoglyphNumbers ConfidenceMap = ConfidenceMap{
	Map: map[rune][][]rune{
		'0': {{'O'}, {'o'}},
		'1': {{'l'}, {'I'}},
		'3': {{'E'}},
		'6': {{'b'}},
	},
	MaxConfidence: 1,
}

// ASCIIHomoglyphSymbols maps ascii symbols to their homoglyphs.
var ASCIIHomoglyphSymbols ConfidenceMap = ConfidenceMap{
	Map: map[rune][][]rune{
		'a': {{'&'}, {'@'}},
		's': {{'$'}},
		'l': {{'|'}},

		'S': {{'$'}},

		'$': {{'S'}, {'s'}},
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
		b, next, err := t.processFirst(ctx, in.Message, confidence)
		if err != nil {
			return in, err
		}

		// FASTPATH: unmodified string.
		if b.Cap() == 0 {
			return in, nil
		}

		// check if string is valid after the first modification.
		out := b.String() + next
		ok, err := t.cfg.Source.Valid(ctx, out)
		if err != nil || ok {
			in.setAndIncrement(out)
			return in, err
		}

		iter := &processNextIterator{
			b:          b,
			t:          t,
			confidence: confidence,
			next:       next,
		}

		for iter.Next() {
			out, err := iter.Value()
			if err != nil {
				return MessagePacket{}, err
			}

			ok, err := t.cfg.Source.Valid(ctx, out)
			if err != nil || ok {
				in.setAndIncrement(out)
				return in, err
			}
		}

		// collect last value.
		out, err = iter.Value()
		if err != nil {
			return MessagePacket{}, err
		}
		ok, err = t.cfg.Source.Valid(ctx, out)
		if err != nil || ok {
			in.setAndIncrement(out)
			return in, err
		}
	}

	return in, nil
}

// processFirst runs a copy on write run through s, allocating a buffer
// with the capacity of maxBytes.
//
// it returns a pointer to the allocated builder and the remaining string (unprocessed) +
// any errors from the context.
func (t *homoglyphTransformer) processFirst(ctx context.Context, s string, confidence int) (*strings.Builder, string, error) {
	for i, c := range s {
		select {
		case <-ctx.Done():
			return nil, "", ctx.Err()
		default:
		}

		rs := t.getRunes(c, confidence)
		if len(rs) == 1 && rs[0] == c && c != utf8.RuneError {
			continue
		}

		var width int
		if c == utf8.RuneError {
			c, width = utf8.DecodeRuneInString(s[i:])
			// genuine utf8.RuneError bytes.
			if len(rs) == 1 && width != 1 && c == rs[0] {
				continue
			}
		} else {
			width = utf8.RuneLen(c)
		}

		var widthT int
		for _, r := range rs {
			widthT += utf8.RuneLen(r)
		}
		if remWidth, widthRem := t.cfg.MaxBytes-len(s[:i]), widthT+len(s[i+width:]); remWidth-widthRem < 0 {
			continue
		}

		var b strings.Builder
		if t.cfg.MaxBytes > 0 {
			b.Grow(t.cfg.MaxBytes)
		} else {
			b.Grow(len(s) + utf8.UTFMax)
		}

		b.WriteString(s[:i])
		for _, r := range rs {
			b.WriteRune(r)
		}
		s = s[i+width:]

		return &b, s, nil
	}

	return &strings.Builder{}, s, nil
}

// getRunes gets the mapped rune for c at the given confidence level.
//
// if the confidence level isnt found or no match is found for c then the
// original value is returned.
func (t *homoglyphTransformer) getRunes(c rune, confidence int) []rune {
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
			return []rune{c}
		}

		return replaceRunes[confidence]
	}

	return []rune{c}
}

// processNextIterator is an iterator for processing rune by rune the remaining
// value from the first processing phase.
//
// processNextIterator makes sure that the maxBytes field is followed.
type processNextIterator struct {
	b          *strings.Builder
	t          *homoglyphTransformer
	confidence int
	next       string
	err        error
}

// Next advances the iterator through the next rune and writes to the
// builder any modification.
func (p *processNextIterator) Next() bool {
	if len(p.next) == 0 || p.err != nil {
		return false
	}

	c, width := utf8.DecodeRuneInString(p.next)
	rs := p.t.getRunes(c, p.confidence)

	// no MaxBytes config field, just add the values.
	if p.t.cfg.MaxBytes <= 0 {
		for _, r := range rs {
			if r < utf8.RuneSelf {
				p.b.WriteByte(byte(r))
			} else {
				p.b.WriteRune(r)
			}
		}
		p.next = p.next[width:]
		return true
	}

	var widthT int
	for _, r := range rs {
		widthT += utf8.RuneLen(r)
	}

	// there must be at least space for p.next bytes in the buffer, we try
	// to determine if there is also space for widthT bytes.
	bufBytes := p.t.cfg.MaxBytes - p.b.Len()
	switch n := bufBytes - len(p.next); {
	case n > 0: // possible space for growth.
		if n+width-widthT >= 0 {
			for _, r := range rs {
				if r < utf8.RuneSelf {
					p.b.WriteByte(byte(r))
				} else {
					p.b.WriteRune(r)
				}
			}
			p.next = p.next[width:]
			return true
		}

		// not enough space for growth. (continue since there still is space)
		if c < utf8.RuneSelf {
			p.b.WriteByte(byte(c))
		} else {
			p.b.WriteRune(c)
		}
		p.next = p.next[width:]
		return true

	case n == 0: // no more space for growth. Accept only for 1:1 matches or lower.
		// substitution width bigger and cant grow, write original value.
		if widthT > width {
			if c < utf8.RuneSelf {
				p.b.WriteByte(byte(c))
			} else {
				p.b.WriteRune(c)
			}
			p.next = p.next[width:]
			return true
		}

		// write smaller or 1:1 matches.
		for _, r := range rs {
			if r < utf8.RuneSelf {
				p.b.WriteByte(byte(r))
			} else {
				p.b.WriteRune(r)
			}
		}
		p.next = p.next[width:]
		return true

	default: // anomaly.
		p.err = errors.New("sinoname: unexpected number of bytes")
		return false
	}
}

// Value builds the end value from the builder.
func (p *processNextIterator) Value() (string, error) {
	out := p.b.String() + p.next
	return out, p.err
}
