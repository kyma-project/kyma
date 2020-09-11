package pretty

type Kind int

const (
	Function Kind = iota
	FunctionType
	Functions
	FunctionsType

	Repository
	RepositoryType
	Repositories
	RepositoriesType
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
	case Repository:
		return "GitRepository"
	case RepositoryType:
		return "GitRepository"
	case Repositories:
		return "GitRepositories"
	case RepositoriesType:
		return "[]GitRepository"
	default:
		return "Function"
	}
}
