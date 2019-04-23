package pretty

type Kind int

const (
	BackendModule Kind = iota
	BackendModules
	Microfrontend
	Microfrontends
)

func (k Kind) String() string {
	switch k {
	case BackendModule:
		return "BackendService"
	case BackendModules:
		return "BackendServices"
	case Microfrontend:
		return "Microfrontend"
	case Microfrontends:
		return "Microfrontends"
	default:
		return ""
	}
}
