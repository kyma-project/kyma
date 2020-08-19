package clusterassetgroup

const (
	KeyOpenApiSpec  = "openapi"
	KeyODataSpec    = "odata"
	KeyAsyncApiSpec = "asyncapi"

	SpecHashFormat = "SpecHash-%s"
)

type Entry struct {
	Id          string
	DisplayName string
	Description string
	Assets      []Asset
	Labels      map[string]string
	Status      StatusType
}

type Asset struct {
	ID       string
	Name     string
	Url      string
	Type     ApiType
	Format   SpecFormat
	SpecHash string
	Content  []byte
}

type StatusType string

const (
	StatusNone StatusType = ""
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
