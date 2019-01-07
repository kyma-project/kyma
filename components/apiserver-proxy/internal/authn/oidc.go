package authn

import (
	"context"

	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/plugin/pkg/authenticator/token/oidc"
)

// OIDCConfig represents configuration used for JWT request authentication
type OIDCConfig struct {
	IssuerURL            string
	ClientID             string
	CAFile               string
	UsernameClaim        string
	UsernamePrefix       string
	GroupsClaim          string
	GroupsPrefix         string
	SupportedSigningAlgs []string
}

type Authenticator struct {
	auth authenticator.Token
}

func (a *Authenticator) AuthenticateRequest(ctx context.Context) (user.Info, bool, error) {
	// TODO: get token from context
	token := ""
	return a.auth.AuthenticateToken(token)
}

// NewOIDCAuthenticator returns OIDC authenticator
func NewOIDCAuthenticator(config *OIDCConfig) (*Authenticator, error) {
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

	return &Authenticator{tokenAuthenticator}, nil
}
