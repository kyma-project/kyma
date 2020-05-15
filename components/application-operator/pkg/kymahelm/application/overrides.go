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
	DomainName                            string `json:"DomainName,omitempty"`
	ApplicationGatewayImage               string `json:"ApplicationGatewayImage,omitempty"`
	ApplicationGatewayTestsImage          string `json:"ApplicationGatewayTestsImage,omitempty"`
	EventServiceImage                     string `json:"EventServiceImage,omitempty"`
	EventServiceTestsImage                string `json:"EventServiceTestsImage,omitempty"`
	ApplicationConnectivityValidatorImage string `json:"ApplicationConnectivityValidatorImage,omitempty"`
	Tenant                                string `json:"Tenant,omitempty"`
	Group                                 string `json:"Group,omitempty"`
	GatewayOncePerNamespace               bool   `json:"GatewayOncePerNamespace,omitempty"`
	StrictMode                            string `json:"StrictMode,omitempty"`
}
