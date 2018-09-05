package pretty

type Kind int

const (
	Function Kind = iota
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
