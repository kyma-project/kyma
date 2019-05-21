package application

const (
	overridesTemplate = `global:
    domainName: {{ .DomainName }}
    applicationGatewayImage: {{ .ApplicationGatewayImage }}
    applicationGatewayTestsImage: {{ .ApplicationGatewayTestsImage }}
    eventServiceImage: {{ .EventServiceImage }}
    eventServiceTestsImage: {{ .EventServiceTestsImage }}
    appConnectorValidationProxyImage: {{ .AppConnectorValidationProxyImage }}
    tenant: {{ .Tenant }}
    group: {{ .Group }}`
)

type OverridesData struct {
	DomainName                       string
	ApplicationGatewayImage          string
	ApplicationGatewayTestsImage     string
	EventServiceImage                string
	EventServiceTestsImage           string
	AppConnectorValidationProxyImage string
	Tenant                           string
	Group                            string
}
