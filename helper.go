package sinoname

import (
	"math/rand"
	"strings"
	"time"
)

func replaceRunes(in string, compare ReplaceMap) string {
	r := rand.New(
		rand.NewSource(time.Now().Unix()),
	)

	out := []rune(in)
	for i, v := range []rune(in) {
		if vals, ok := compare[v]; ok {
			out[i] = vals[r.Intn(len(vals))]
		}
	}

	return string(out)
}

func splitOnSpecial(in string) string {
	r := strings.NewReplacer(".", " ", "_", " ", "-", " ")
	return r.Replace(strings.TrimSpace(in))
}
