package testkit

type ConnectRequest struct {
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

type API struct {
	ID          string `json:"id"`
	Provider    string `json:"provider"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
