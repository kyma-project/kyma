package gateway

const (
	overridesTemplate = `global:
    applicationGatewayImage: {{ .ApplicationGatewayImage }}
    applicationGatewayTestsImage: {{ .ApplicationGatewayTestsImage }}
    deployGatewayOncePerNamespace: {{ .GatewayOncePerNamespace }}`
)

type OverridesData struct {
	ApplicationGatewayImage      string
	ApplicationGatewayTestsImage string
	GatewayOncePerNamespace      bool
}
