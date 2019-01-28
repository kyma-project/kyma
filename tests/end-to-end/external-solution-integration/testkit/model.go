package testkit

type ConnectRequest struct {
	IsLocalKyma        bool   `json:"localKyma"`
	URL                string `json:"url"`
	ShouldRegisterAPIs bool   `json:"register"`
	MockHostname       string `json:"hostname"`
}

type ConnectResponse struct {
	MetadataURL   string `json:"metadataUrl"`
	EventsURL     string `json:"eventsUrl"`
	ClusterDomain string `json:"cluster_domain"`
	AppName       string `json:"re_name"`
}
