package config

// Source is an interface which should be implemented by the client, it serves the purpose
// of validating a username and should be concurrency safe.
//
// You can implement source on your database instance to validate usernames.
type Source interface {
	// Valid takes in a usernames and checks wether its valid or not, if the username is valid
	// the return values should be true, nil.
	//
	// If the username already exists the return value should be false, nil.
	//
	// If there is an error with the source the return value should be false, err.
	// The whole pipeline is closed if the source returns an error, rendering it unreliable.
	Valid(string) (bool, error)
}
