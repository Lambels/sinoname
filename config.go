package sinoname

// Config represents a config object accepted by sinoname.New() and sinoname.LayerFactor() .
type Config struct {
	// MaxLen is used to set the max length of the returned value.
	MaxLen int

	// MaxVals is used to set the max number of variations of the initial input, if the max
	// value cant be reached the max number of possible variations is returned.
	MaxVals int

	// PreventDefault prevents the default value from being recovered.
	PreventDefault bool

	// Fallback is used to check if the raw initial value should pass through all the stages
	// or if the value should be returned as soon as possible.
	Fallback bool

	// Source is used to validate if the products of the shuffle are unique / valid.
	Source Source
}
