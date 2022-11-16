package sinoname

import (
	"math/rand"
	"sync"
)

// Config represents a config object accepted by sinoname.New(), sinoname.LayerFactor and sinoname.TransformerFactory .
type Config struct {
	// MaxLen is used to set the max length of the returned values (in bytes),
	// it must be enforced by the transformers to make sure that they dont return
	// values longer then MaxLen
	MaxLen int

	// MaxVals is used to set the max number of returned values.
	// The consumer reads up to MaxVals values.
	MaxVals int

	// PreventDefault prevents the default value from being read by the consumer.
	PreventDuplicates bool

	// Source is used to validate if the products of the transformers are unique / valid.
	Source Source

	// SplitOn is a slice of symbols used by the case transformers (camel case, kebab case, ...)
	// to decide where to split the word up and add their specific separator.
	SplitOn []string

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
