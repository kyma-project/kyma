package pretty

type Kind int

const (
	BackendModule Kind = iota
	BackendModules
	MicroFrontend
	MicroFrontends
	ClusterMicroFrontend
	ClusterMicroFrontends
	NavigationNode
)

func (k Kind) String() string {
	switch k {
	case BackendModule:
		return "BackendService"
	case BackendModules:
		return "BackendServices"
	case MicroFrontend:
		return "MicroFrontend"
	case MicroFrontends:
		return "MicroFrontends"
	case ClusterMicroFrontend:
		return "ClusterMicroFrontend"
	case ClusterMicroFrontends:
		return "ClusterMicroFrontends"
	case NavigationNode:
		return "NavigationNode"
	default:
		return ""
	}
}
