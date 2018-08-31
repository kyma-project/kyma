package pretty

type Kind int

const (
	Unknown Kind = iota
	ApiSpec
	AsyncApiSpec
	Content
	Topic
	Topics
)

func (k Kind) String() string {
	switch k {
	case ApiSpec:
		return "API"
	case AsyncApiSpec:
		return "AsyncAPI"
	case Content:
		return "Content"
	case Topic:
		return "Topic"
	case Topics:
		return "Topics"
	default:
		return ""
	}
}
