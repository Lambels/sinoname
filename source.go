package sinoname

type Source interface {
	Valid(string) (bool, error)
}
