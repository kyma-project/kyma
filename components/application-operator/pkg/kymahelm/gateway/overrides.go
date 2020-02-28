package gateway

const (
	overridesTemplate = `global:
    applicationGatewayImage: {{ .ApplicationGatewayImage }}
    applicationGatewayTestsImage: {{ .ApplicationGatewayTestsImage }}`
)

type OverridesData struct {
	ApplicationGatewayImage      string
	ApplicationGatewayTestsImage string
}
