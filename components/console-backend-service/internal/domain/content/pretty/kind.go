package pretty

type Kind int

const (
	ApiSpec Kind = iota
	OpenApiSpec
	ODataSpec
	AsyncApiSpec
	Content
	Topic
	Topics
)

func (k Kind) String() string {
	switch k {
	case ApiSpec:
		return "API"
	case OpenApiSpec:
		return "OpenApi"
	case ODataSpec:
		return "OData"
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
