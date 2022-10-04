package sinoname

import (
	"strings"
)

func SplitOnSpecial(in string) []string {
	r := strings.NewReplacer(".", " ", "_", " ", "-", " ")
	v := r.Replace(strings.TrimSpace(in))

	return strings.Fields(v)
}
