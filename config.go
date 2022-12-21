package sinoname

import (
	"math/rand"
	"sync"
)

// Config represents a config object accepted by sinoname.New(), sinoname.LayerFactor and sinoname.TransformerFactory .
type Config struct {
	// MaxBytes is used to set the max length of the returned values (in bytes),
	// it must be enforced by the transformers to make sure that they dont return
	// values longer then MaxBytes
	MaxBytes int

	// MaxVals is used to set the max number of returned values.
	// The consumer reads up to MaxVals values.
	MaxVals int

	// MaxChanges is used to determine the max number of changes which can be
	// done on a message.
	//
	// Each modification by a transformer is marked as a change.
	MaxChanges int

	// PreventDefault prevents the default value from being read by the consumer.
	PreventDefault bool

	// PreventDefault prevents duplicate values from being read by the consumer.
	PreventDuplicates bool

	// Source is used to validate if the products of the transformers are unique / valid.
	Source Source

	// Tokenize takes in a string and forms tokens from the string.
	// If Tokenize isnt provided, the default tokenize function is used.
	//
	// The Tokenize function is used by transformers like:
	// CamelCase, SnakeCase, ..., RandomOrder.
	Tokenize func(string) []string

	// StripNumbers takes in a string and splits the string into two strings:
	// string containing the letters, string containing the string representation of the numbers.
	// If StripNumbers isnt provided, the defualt stripNumbersASCII is used.
	StripNumbers func(string) (string, string)

	// Adjectives is a slice of adjectives to be used by suffix, prefix and circumfix transformers.
	// Should be shuffled before referenced.
	Adjectives []string

	// RandSrc is used for random opperations throughout the pipeline.
	RandSrc *rand.Rand

	// shuffle pool is non-nil if adjectives are provided, it keeps alive fixed sized
	// buffers of shuffled integers used to shuffle the adjectives slice.
	shufflePool sync.Pool
}

func (c *Config) getShuffle() []int {
	if c.Adjectives == nil {
		return nil
	}

	v, _ := c.shufflePool.Get().([]int)
	return v
}

func (c *Config) putShuffle(slc []int) {
	if c.Adjectives == nil {
		return
	}

	c.RandSrc.Shuffle(len(slc), func(i, j int) { slc[i], slc[j] = slc[j], slc[i] })
	c.shufflePool.Put(slc)
}
