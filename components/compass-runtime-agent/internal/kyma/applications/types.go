package applications

const (
	CredentialsOAuthType = "OAuth"
	CredentialsBasicType = "Basic"
)

type Credentials struct {
	Type              string
	SecretName        string
	AuthenticationUrl string
	CSRFInfo          *CSRFInfo
}

type CSRFInfo struct {
	TokenEndpointURL string
}
