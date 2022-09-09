package layer

import "errors"

type errTransformer struct{}

func (t errTransformer) Transform(string) (string, error) {
	return "", errors.New("test error")
}
