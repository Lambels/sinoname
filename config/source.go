package config

type Source interface {
	Valid(string) (bool, error)
}
