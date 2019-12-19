package clusterassetgroup

const (
	KeyOpenApiSpec  = "openapi"
	KeyODataSpec    = "odata"
	KeyAsyncApiSpec = "asyncapi"

	SpecHash = "SpecHash"
)

type Entry struct {
	Id          string
	DisplayName string
	Description string
	Urls        map[string]string
	Labels      map[string]string
	SpecHash    string
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
	AsyncApi     ApiType = "asyncapi"
	Empty        ApiType = ""
)

type SpecFormat string

const (
	SpecFormatJSON SpecFormat = "json"
	SpecFormatYAML SpecFormat = "yaml"
	SpecFormatXML  SpecFormat = "xml"
)

type SpecCategory string

const (
	ApiSpec      SpecCategory = "apispec"
	EventApiSpec SpecCategory = "eventapispec"
)
