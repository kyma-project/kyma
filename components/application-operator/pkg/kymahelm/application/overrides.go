package application

const (
	overridesTemplate = `global:
    domainName: {{ .DomainName }}
    applicationGatewayImage: {{ .ApplicationGatewayImage }}
    applicationGatewayTestsImage: {{ .ApplicationGatewayTestsImage }}
    eventServiceImage: {{ .EventServiceImage }}
    eventServiceTestsImage: {{ .EventServiceTestsImage }}
    applicationConnectivityValidatorImage: {{ .ApplicationConnectivityValidatorImage }}
    tenant: {{ .Tenant }}
    group: {{ .Group }}
    deployGatewayOncePerNamespace: {{ .GatewayOncePerNamespace }}
    strictMode: {{ .StrictMode }}`
)

type OverridesData struct {
	DomainName                            string
	ApplicationGatewayImage               string
	ApplicationGatewayTestsImage          string
	EventServiceImage                     string
	EventServiceTestsImage                string
	ApplicationConnectivityValidatorImage string
	Tenant                                string
	Group                                 string
	GatewayOncePerNamespace               bool
	StrictMode                            string
}
