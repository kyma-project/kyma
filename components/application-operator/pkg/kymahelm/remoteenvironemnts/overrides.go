package remoteenvironemnts

const (
	OverridesTemplate = `global:
  domainName: {{ .DomainName }}
  proxyServiceImage: {{ .ProxyServiceImage }}
  eventServiceImage: {{ .EventServiceImage }}
  eventServiceTestsImage: {{ .EventServiceTestsImage }}`
)

type OverridesData struct {
	DomainName             string
	ProxyServiceImage      string
	EventServiceImage      string
	EventServiceTestsImage string
}
