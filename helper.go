package sinoname

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	capital = 1
	other   = 2
)

func tokenizeDefault(in string) []string {
	if !utf8.ValidString(in) || in == "" {
		return []string{in}
	}

	// split on symbols first.
	r := strings.NewReplacer(splitOnDefault...)
	v := r.Replace(strings.TrimSpace(in))
	split := strings.Fields(v)
	var runes [][]rune
	var out []string

	// split on camel case.
	for _, field := range split {
		var lastClass int
		var class int
		var lastNumbers int

		for i, r := range field {
			switch {
			case unicode.IsUpper(r):
				class = capital

			case unicode.IsNumber(r):
				lastNumbers++
				continue

			default:
				class = other
			}

			// split on numbers.
			if lastNumbers > 1 {
				split := field[i-lastNumbers : i]
				runes = append(runes, []rune(split))

				// trigger a split regardless.
				class = -1
				lastNumbers = 0 // lastNumbers accounted for.
			}

			width := utf8.RuneLen(r)
			if class != lastClass {
				runes = append(runes, []rune(field[i-lastNumbers:i+width]))
			} else {
				runes[len(runes)-1] = append(runes[len(runes)-1], []rune(field[i-lastNumbers:i+width])...)
			}

			lastNumbers = 0
			lastClass = class
		}

		// handle trailing numbers.
		if lastNumbers > 1 { // create separate field.
			split := field[len(field)-lastNumbers:]
			runes = append(runes, []rune(split))
		} else if lastNumbers > 0 { // attach number at trailing field.
			runes[len(runes)-1] = append(runes[len(runes)-1], rune(field[len(field)-1]))
		}
	}

	// stick back runes.
	for i := 0; i < len(runes)-1; i++ {
		if unicode.IsUpper(runes[i][len(runes[i])-1]) && unicode.IsLower(runes[i+1][0]) {
			// posibility for letter to have number next to it. (only one)
			var offset int
			if len(runes[i]) > 1 && unicode.IsNumber(runes[i][len(runes[i])-2]) {
				offset = 1
			}

			runes[i+1] = append(runes[i][len(runes[i])-1-offset:len(runes[i])], runes[i+1]...)
			runes[i] = runes[i][:len(runes[i])-1-offset]
		}
	}

	for _, r := range runes {
		if len(r) > 0 {
			out = append(out, string(r))
		}
	}

	return out
}

func stripNumbersASCII(in string) (string, string) {
	var bL strings.Builder
	var bD strings.Builder
	for _, r := range in {
		switch {
		case r >= 48 && r <= 57:
			bD.WriteRune(r)
		default:
			bL.WriteRune(r)
		}
	}

	return bL.String(), bD.String()
}

func StripNumbersUnicode(in string) (string, string) {
	var bL strings.Builder
	var bD strings.Builder
	for _, r := range in {
		switch {
		case unicode.IsDigit(r):
			bD.WriteRune(r)
		default:
			bL.WriteRune(r)
		}
	}

	return bL.String(), bD.String()
}
