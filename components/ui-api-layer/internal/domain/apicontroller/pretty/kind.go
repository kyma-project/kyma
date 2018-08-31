package pretty

type Kind int

const (
	Unknown Kind = iota
	API
	APIs
)

func (k Kind) String() string {
	switch k {
	case API:
		return "API"
	case APIs:
		return "APIs"
	default:
		return ""
	}
}
