package application

const (
	OverridesTemplate = `global:
  domainName: {{ .DomainName }}
  applicationProxyImage: {{ .ApplicationProxyImage }}
  eventServiceImage: {{ .EventServiceImage }}
  eventServiceTestsImage: {{ .EventServiceTestsImage }}`
)

type OverridesData struct {
	DomainName             string
	ApplicationProxyImage  string
	EventServiceImage      string
	EventServiceTestsImage string
}
