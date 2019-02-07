package pretty

type Kind int

const (
	EventActivation Kind = iota
	EventActivations
	EventActivationEvents
	Application
	Applications
	ConnectorService
	Namespace
	Namespaces
	ApplicationMapping
)

func (k Kind) String() string {
	switch k {
	case EventActivation:
		return "Event Activation"
	case EventActivations:
		return "Event Activations"
	case EventActivationEvents:
		return "Event Activation Events"
	case Application:
		return "Application"
	case Applications:
		return "Applications"
	case ConnectorService:
		return "Connector Service"
	case Namespace:
		return "Namespace"
	case Namespaces:
		return "Namespaces"
	case ApplicationMapping:
		return "ApplicationMapping"
	default:
		return ""
	}
}
