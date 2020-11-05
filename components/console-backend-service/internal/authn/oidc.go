package authn

import (
	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/authentication/request/bearertoken"
	"k8s.io/apiserver/plugin/pkg/authenticator/token/oidc"
)

// OIDCConfig represents configuration used for JWT request authentication
type OIDCConfig struct {
	IssuerURL            string   `envconfig:"default=https://dex.kyma.local"`
	ClientID             string   `envconfig:"default=kyma-client"`
	CAFile               string   `envconfig:"optional"`
	UsernameClaim        string   `envconfig:"default=email"`
	UsernamePrefix       string   `envconfig:"optional"`
	GroupsClaim          string   `envconfig:"default=groups"`
	GroupsPrefix         string   `envconfig:"optional"`
	SupportedSigningAlgs []string `envconfig:"default=RS256"`
}

// NewOIDCAuthenticator returns OIDC authenticator
func NewOIDCAuthenticator(config *OIDCConfig) (authenticator.Request, error) {
	tokenAuthenticator, err := oidc.New(oidc.Options{
		IssuerURL:            config.IssuerURL,
		ClientID:             config.ClientID,
		CAFile:               config.CAFile,
		UsernameClaim:        config.UsernameClaim,
		UsernamePrefix:       config.UsernamePrefix,
		GroupsClaim:          config.GroupsClaim,
		GroupsPrefix:         config.GroupsPrefix,
		SupportedSigningAlgs: config.SupportedSigningAlgs,
	})
	if err != nil {
		return nil, err
	}

	return bearertoken.New(tokenAuthenticator), nil
}
