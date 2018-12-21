package pretty

type Kind int

const (
	BackendModule Kind = iota
	BackendModules Kind = iota
)

func (k Kind) String() string {
	switch k {
	case BackendModule:
		return "BackendService"
	case BackendModules:
		return "BackendServices"
	default:
		return ""
	}
}
