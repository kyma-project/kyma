package application

type OverridesData struct {
	DomainName                            string `json:"domainName,omitempty"`
	ApplicationGatewayImage               string `json:"applicationGatewayImage,omitempty"`
	ApplicationGatewayTestsImage          string `json:"applicationGatewayTestsImage,omitempty"`
	EventServiceImage                     string `json:"eventServiceImage,omitempty"`
	EventServiceTestsImage                string `json:"eventServiceTestsImage,omitempty"`
	ApplicationConnectivityValidatorImage string `json:"applicationConnectivityValidatorImage,omitempty"`
	Tenant                                string `json:"tenant,omitempty"`
	Group                                 string `json:"group,omitempty"`
	LogLevel                              string `json:"logLevel,omitempty"`
	LogFormat                             string `json:"logFormat,omitempty"`
	GatewayOncePerNamespace               bool   `json:"deployGatewayOncePerNamespace,omitempty"`
	StrictMode                            string `json:"strictMode,omitempty"`
	IsBEBEnabled                          bool   `json:"isBEBEnabled,omitempty"`
}
