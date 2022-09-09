package transformer

type Transformer interface {
	Transform(in string) (string, error)
}

type Signal struct {
	Err error
	Val string
}

func TransformWithSignal(t Transformer, val string) <-chan Signal {
	ch := make(chan Signal)

	go func() {
		val, err := t.Transform(val)
		ch <- Signal{
			Err: err,
			Val: val,
		}
	}()

	return ch
}
