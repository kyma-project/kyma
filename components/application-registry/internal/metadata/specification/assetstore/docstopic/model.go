package docstopic

const (
	KeyOpenApiSpec       = "openapi"
	KeyODataXMLSpec      = "odataxml"
	KeyODataJSONSpec     = "odatajson"
	KeyEventsSpec        = "events"
	KeyDocumentationSpec = "docs"
)

type Entry struct {
	Id          string
	DisplayName string
	Description string
	Urls        map[string]string
	Labels      map[string]string
}

type ApiType string

const (
	OpenApiType  ApiType = "openapi"
	ODataApiType ApiType = "odata"
	NoneApiType  ApiType = ""
)
