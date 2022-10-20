package sinoname

import (
	"strings"
)

func SplitOnSpecial(in string, special []string) []string {
	r := strings.NewReplacer(special...)
	v := r.Replace(strings.TrimSpace(in))

	return strings.Fields(v)
}

func StripNumbers(in string) (string, string) {
	var bL strings.Builder
	var bD strings.Builder
	for _, v := range in {
		switch {
		case v >= 48 && v <= 57:
			bD.WriteRune(v)
		default:
			bL.WriteRune(v)
		}
	}

	return bL.String(), bD.String()
}
