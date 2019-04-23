package pretty

type Kind int

const (
	BackendModule Kind = iota
	BackendModules
	Microfrontend
	Microfrontends
	ClusterMicrofrontend
	ClusterMicrofrontends
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
	case ClusterMicrofrontend:
		return "ClusterMicrofrontend"
	case ClusterMicrofrontends:
		return "ClusterMicrofrontends"
	default:
		return ""
	}
}
