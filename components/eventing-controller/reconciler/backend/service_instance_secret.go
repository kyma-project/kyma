package backend

type Message struct {
	Broker Broker           `json:"broker,omitempty"`
	OA2    OAuthCredentials `json:"oa2,omitempty"`
	URI    string           `json:"uri,omitempty"`
}

type Broker struct {
	BrokerType string `json:"type,omitempty"`
}

type OAuthCredentials struct {
	ClientID      string `json:"clientid,omitempty"`
	ClientSecret  string `json:"clientsecret,omitempty"`
	GrantType     string `json:"granttype,omitempty"`
	TokenEndpoint string `json:"tokenendpoint,omitempty"`
}
