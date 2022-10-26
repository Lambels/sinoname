package sinoname

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
	PreventDefault bool

	// Source is used to validate if the products of the transformers are unique / valid.
	Source Source

	// SplitOn is a slice of symbols used by the case transformers (camel case, kebab case, ...)
	// to decide where to split the word up and add their specific separator.
	SplitOn []string

	SingleHomoglyphTables []ConfidenceMap

	// MultiHomoglyphTables []map[rune][]rune
}
