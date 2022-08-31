package helper

import (
	"math/rand"
	"strings"
	"time"
)

type ReplaceMap map[rune][]rune

func ReplaceRunes(in string, compare ReplaceMap) string {
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

func SplitOnSpecial(in string) string {
	r := strings.NewReplacer(".", " ", "_", " ", "-", " ")
	return r.Replace(strings.TrimSpace(in))
}
