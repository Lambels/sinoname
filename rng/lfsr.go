package rng

import (
	"time"
)

var _ PRNG = (*Lfsr8)(nil)

// Lfsr8 represents an 8 bit linear feedback shift register
type Lfsr8 struct {
	state uint8
	seed  uint8
}

// NewLfsr8 returns a linear feedback shift register initialized with the specified seed. If the seed is zero the seed is initialized using the current time.
func NewLfsr8(seed uint) *Lfsr8 {
	l := &Lfsr8{}
	l.Seed(int(seed))
	return l
}

// Next returns the next pseudo random number from the linear feedback shift register and the restarted flag
// which indicates that the sequence has completed and is restarting.
func (l *Lfsr8) Next() (value int, restarted bool) {
	s := l.state
	b := (s >> 0) ^ (s >> 2) ^ (s >> 3) ^ (s >> 4)
	l.state = (s >> 1) | (b << 7)
	return int(l.state), l.state == l.seed
}

// Seed re seeds the Lsfr generator.
func (l *Lfsr8) Seed(seed int) {
	// truncate bits to uint8 value and then check for 0 value.
	for seed&0xff == 0 {
		seed = int(time.Now().Nanosecond() & 0xff)
	}

	l.seed = uint8(seed)
	l.state = uint8(seed)
}

// Range returns the range of values the prng can generate.
func (l *Lfsr8) Range() int {
	return 0xff
}

var _ PRNG = (*Lfsr16)(nil)

// Lfsr16 represents an 16 bit linear feedback shift register
type Lfsr16 struct {
	state uint16
	seed  uint16
}

// NewLfsr16 returns a linear feedback shift register initialized with the specified seed. If the seed is zero the seed is initialized using the current time.
func NewLfsr16(seed uint) *Lfsr16 {
	l := &Lfsr16{}
	l.Seed(int(seed))
	return l
}

// Next returns the next pseudo random number from the linear feedback shift register and the restarted flag
// which indicates that the sequence has completed and is restarting.
func (l *Lfsr16) Next() (value int, restarted bool) {
	s := l.state
	b := (s >> 0) ^ (s >> 2) ^ (s >> 3) ^ (s >> 5)
	l.state = (s >> 1) | (b << 15)
	return int(l.state), l.state == l.seed
}

// Seed re seeds the Lsfr generator.
func (l *Lfsr16) Seed(seed int) {
	// truncate bits to uint16 value and then check for 0 value.
	for seed&0xffff == 0 {
		seed = int(time.Now().Nanosecond() & 0xffff)
	}

	l.seed = uint16(seed)
	l.state = uint16(seed)
}

// Range returns the range of values the prng can generate.
func (l *Lfsr16) Range() int {
	return 0xffff
}
