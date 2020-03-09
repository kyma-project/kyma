package eventing

type Kind int

const (
	Trigger Kind = iota
	TriggerType
)

func (k Kind) String() string {
	switch k {
	case Trigger:
		return "Trigger"
	case TriggerType:
		return "Trigger"
	default:
		return ""
	}
}