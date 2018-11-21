package pretty

type Kind int

const (
	EventActivation Kind = iota
	EventActivations
	EventActivationEvents
	RemoteEnvironment
	RemoteEnvironments
	ConnectorService
	Environment
	Environments
	EnvironmentMapping
)

func (k Kind) String() string {
	switch k {
	case EventActivation:
		return "Event Activation"
	case EventActivations:
		return "Event Activations"
	case EventActivationEvents:
		return "Event Activation Events"
	case RemoteEnvironment:
		return "Remote Environment"
	case RemoteEnvironments:
		return "Remote Environments"
	case ConnectorService:
		return "Connector Service"
	case Environment:
		return "Environment"
	case Environments:
		return "Environments"
	case EnvironmentMapping:
		return "EnvironmentMapping"
	default:
		return ""
	}
}
