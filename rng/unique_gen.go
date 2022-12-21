package rng

import (
	"math/rand"
)

var _ PRNG = (*UniqueRangeGen)(nil)

// UniqueRangeGen generate pseudo-random numbers from the provided rand.Rand uniquely
// in the range [0, n).
type UniqueRangeGen struct {
	vals map[int]struct{}
	n    int
	src  *rand.Rand
}

func NewUniqueRangeGen(src *rand.Rand, n int) *UniqueRangeGen {
	return &UniqueRangeGen{
		vals: make(map[int]struct{}),
		n:    n,
		src:  src,
	}
}

// Next advances the generator and generates a new number.
func (g *UniqueRangeGen) Next() (int, bool) {
	if len(g.vals) == g.n {
		return 0, true
	}

	r := g.src.Intn(g.n)
	for _, ok := g.vals[r]; !ok; {
		r = g.src.Intn(g.n)
	}

	return r, len(g.vals) == g.n
}

// Seed re seeds the source.
func (g *UniqueRangeGen) Seed(seed int) {
	g.src.Seed(int64(seed))
}

// Range outputs n. [0, n)
func (g *UniqueRangeGen) Range() int {
	return g.n
}
