package pretty

type Kind int

const (
	Unknown Kind = iota
	EventActivation
	EventActivations
	EventActivationEvents
	RemoteEnvironment
	RemoteEnvironments
	ConnectorService
	Environment
	Environments
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
	default:
		return ""
	}
}
