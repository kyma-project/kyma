package gateway

type OverridesData struct {
	ApplicationGatewayImage      string `json:"ApplicationGatewayImage,omitempty"`
	ApplicationGatewayTestsImage string `json:"ApplicationGatewayTestsImage,omitempty"`
	GatewayOncePerNamespace      bool   `json:"GatewayOncePerNamespace,omitempty"`
}
