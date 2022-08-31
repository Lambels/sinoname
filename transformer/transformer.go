package transformer

type Transformer interface {
	Transform(in string) (string, error)
}
