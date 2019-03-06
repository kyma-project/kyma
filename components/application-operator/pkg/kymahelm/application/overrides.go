package application

const (
	overridesTemplate = `global:
    domainName: {{ .DomainName }}
    applicationGatewayImage: {{ .ApplicationGatewayImage }}
    eventServiceImage: {{ .EventServiceImage }}
    eventServiceTestsImage: {{ .EventServiceTestsImage }}
    subjectCN: {{ .SubjectCN }}`
)

type OverridesData struct {
	DomainName              string
	ApplicationGatewayImage string
	EventServiceImage       string
	EventServiceTestsImage  string
	SubjectCN               string
}
