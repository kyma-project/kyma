package serverless

type Kind string

const (
	KindFunction  Kind = "Function"
	KindFunctions Kind = "Functions"
)

func (k Kind) String() string {
	return string(k)
}
