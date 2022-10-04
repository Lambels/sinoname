package sinoname

import (
	"strings"
)

func SplitOnSpecial(in string, special []string) []string {
	r := strings.NewReplacer(special...)
	v := r.Replace(strings.TrimSpace(in))

	return strings.Fields(v)
}
