package docstopic

const (
	KeyOpenApiSpec       = "openapi"
	KeyODataSpec         = "odata"
	KeyAsyncApiSpec      = "asyncapi"
	KeyDocumentationSpec = "docs"

	ApiSpec       = "ApiSpec"
	Documentation = "Documentation"
	EventsSpec    = "EventsSpec"
)

type Entry struct {
	Id          string
	DisplayName string
	Description string
	Urls        map[string]string
	Labels      map[string]string
	Hashes      map[string]string
	Status      StatusType
}

type StatusType string

const (
	StatusNone    StatusType = ""
	StatusPending StatusType = "Pending"
	StatusFailed  StatusType = "Failed"
	StatusReady   StatusType = "Ready"
)

type ApiType string

const (
	OpenApiType  ApiType = "openapi"
	ODataApiType ApiType = "odata"
	NoneApiType  ApiType = ""
)
