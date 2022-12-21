package rng

// PRNG is an interface which abstracts a PRNG with a custom range.
//
// PRNG is used interchangeably to process higher or lower ranges of values in an
// efficient manner.
type PRNG interface {
	// Next generates the next number and indicates that the range has been reached.
	Next() (int, bool)
	// Seed re seeds the value.
	Seed(int)
	// Range gives the PRNG range.
	Range() int
}
