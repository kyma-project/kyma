package pretty

type Kind int

const (
	Unknown Kind = iota
	Function
	Functions
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
