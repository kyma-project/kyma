package application

const (
	overridesTemplate = `global:
    domainName: {{ .DomainName }}
    applicationGatewayImage: {{ .ApplicationGatewayImage }}
    applicationGatewayTestsImage: {{ .ApplicationGatewayTestsImage }}
    eventServiceImage: {{ .EventServiceImage }}
    eventServiceTestsImage: {{ .EventServiceTestsImage }}
    subjectCN: {{ .SubjectCN }}
    subjectOrganization: {{ .SubjectOrganization }}
    subjectOrganizationUnit: {{ .SubjectOrganizationUnit }}`
)

type OverridesData struct {
	DomainName                   string
	ApplicationGatewayImage      string
	ApplicationGatewayTestsImage string
	EventServiceImage            string
	EventServiceTestsImage       string
	SubjectCN                    string
	SubjectOrganization          string
	SubjectOrganizationUnit      string
}
