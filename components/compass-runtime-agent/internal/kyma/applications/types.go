package applications

type Credentials struct {
	Type              string
	SecretName        string
	AuthenticationUrl string
	CSRFInfo          *CSRFInfo
	Headers           *map[string][]string
	QueryParameters   *map[string][]string
}

type CSRFInfo struct {
	TokenEndpointURL string
}
