package pretty

type Kind string

const (
	Function  Kind = "Function"
	Functions Kind = "Functions"
)

func (k Kind) String() string {
	switch k {
	case Function:
		return "Function"
	case Functions:
		return "Functions"
	default:
		return ""
	}
}
