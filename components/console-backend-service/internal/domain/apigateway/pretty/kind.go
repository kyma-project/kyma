package pretty

type Kind int

const (
	APIRule Kind = iota
	APIRules
)

func (k Kind) String() string {
	switch k {
	case APIRule:
		return "APIRule"
	case APIRules:
		return "APIRules"
	default:
		return ""
	}
}
