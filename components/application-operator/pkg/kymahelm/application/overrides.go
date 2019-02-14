package application

const (
	OverridesTemplate = `global:
    domainName: {{ .DomainName }}
    applicationProxyImage: {{ .ApplicationProxyImage }}
    eventServiceImage: {{ .EventServiceImage }}
    eventServiceTestsImage: {{ .EventServiceTestsImage }}
    tenant: {{ .Tenant }}
    group: {{ .Group }}`
)

type OverridesData struct {
	DomainName             string
	ApplicationProxyImage  string
	EventServiceImage      string
	EventServiceTestsImage string
	Tenant                 string
	Group                  string
}
