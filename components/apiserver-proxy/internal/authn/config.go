package authn

// AuthnHeaderConfig contains authentication header settings which enable more information about the user identity to be sent to the upstream
type AuthnHeaderConfig struct {
	// When set to true, kube-rbac-proxy adds auth-related fields to the headers of http requests sent to the upstream
	Enabled bool
	// Corresponds to the name of the field inside a http(2) request header
	// to tell the upstream server about the user's name
	UserFieldName string
	// Corresponds to the name of the field inside a http(2) request header
	// to tell the upstream server about the user's groups
	GroupsFieldName string
	// The separator string used for concatenating multiple group names in a groups header field's value
	GroupSeparator string
}

// AuthnConfig holds all configurations related to authentication options 
type AuthnConfig struct {
	X509   *X509Config
	Header *AuthnHeaderConfig
	OIDC   *OIDCConfig
}

// X509Config holds public client certificate used for authentication requests if specified
type X509Config struct {
	ClientCAFile string
}