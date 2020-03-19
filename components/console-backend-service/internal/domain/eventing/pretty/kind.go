package pretty

type Kind int

const (
	Trigger Kind = iota
	Triggers
	TriggerType
)

func (k Kind) String() string {
	switch k {
	case Trigger:
		return "Trigger"
	case Triggers:
		return "Triggers"
	case TriggerType:
		return "Trigger"
	default:
		return ""
	}
}
