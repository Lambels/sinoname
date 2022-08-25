package sinoname

type ReplaceMap map[rune][]rune

type Transformer interface {
	Transform(in string) (string, error)
}
