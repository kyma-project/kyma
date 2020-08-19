package pretty

type Kind int

const (
	Function Kind = iota
	FunctionType
	Functions
	FunctionsType
)

func (k Kind) String() string {
	switch k {
	case Function:
		return "Function"
	case FunctionType:
		return "Function"
	case Functions:
		return "Functions"
	case FunctionsType:
		return "[]Function"
	default:
		return "Function"
	}
}
