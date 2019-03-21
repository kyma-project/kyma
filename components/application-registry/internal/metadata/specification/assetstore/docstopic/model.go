package docstopic

const (
	DocsTopicKeyOpenApiSpec       = "openapi"
	DocsTopicKeyODataXMLSpec      = "odataxml"
	DocsTopicKeyODataJSONSpec     = "odatajson"
	DocsTopicKeyEventsSpec        = "events"
	DocsTopicKeyDocumentationSpec = "docs"
)

type Entry struct {
	Id          string
	DisplayName string
	Description string
	Urls        map[string]string
	Labels      map[string]string
}
