package application

const (
	overridesTemplate = `global:
    domainName: {{ .DomainName }}
    applicationGatewayImage: {{ .ApplicationGatewayImage }}
    applicationGatewayTestsImage: {{ .ApplicationGatewayTestsImage }}
    eventServiceImage: {{ .EventServiceImage }}
    eventServiceTestsImage: {{ .EventServiceTestsImage }}
    ingressValidationRule: {{ .IngressValidationRule }}`
)

type OverridesData struct {
	DomainName                   string
	ApplicationGatewayImage      string
	ApplicationGatewayTestsImage string
	EventServiceImage            string
	EventServiceTestsImage       string
	IngressValidationRule        string
}
