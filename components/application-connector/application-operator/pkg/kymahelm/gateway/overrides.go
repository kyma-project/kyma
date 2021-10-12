package gateway

type OverridesData struct {
	ApplicationGatewayImage      string `json:"applicationGatewayImage,omitempty"`
	ApplicationGatewayTestsImage string `json:"applicationGatewayTestsImage,omitempty"`
	GatewayOncePerNamespace      bool   `json:"deployGatewayOncePerNamespace,omitempty"`
}
