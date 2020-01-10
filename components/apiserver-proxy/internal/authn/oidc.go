package authn

import (
	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/authentication/request/bearertoken"
	"k8s.io/apiserver/plugin/pkg/authenticator/token/oidc"
)

// OIDCConfig represents configuration used for JWT request authentication
type OIDCConfig struct {
	IssuerURL            string
	ClientIDs            []string
	CAFile               string
	UsernameClaim        string
	UsernamePrefix       string
	GroupsClaim          string
	GroupsPrefix         string
	SupportedSigningAlgs []string
}

//Extends authenticator.Request interface with Cancel() function used to stop underlying authenticator instance once it's not needed anymore
type CancelableAuthRequest interface {
	authenticator.Request
	Cancel() //Cancels (stops) the underlying instance
}

type cancelableAuthRequest struct {
	*bearertoken.Authenticator
	cancelFunc func()
}

func (car *cancelableAuthRequest) Cancel() {
	if car.cancelFunc != nil {
		car.cancelFunc()
	}
}

// NewOIDCAuthenticator returns OIDC authenticator wrapped as a CancelableAuthRequest instance.
// CancelableAuthRequest allows users to cancel the authenticator once it's not used anymore.
func NewOIDCAuthenticator(config *OIDCConfig, clientIndex int) (CancelableAuthRequest, error) {
	tokenAuthenticator, err := oidc.New(oidc.Options{
		IssuerURL:            config.IssuerURL,
		ClientID:             config.ClientIDs[clientIndex],
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

	athntctr := bearertoken.New(tokenAuthenticator)

	return &cancelableAuthRequest{
		Authenticator: athntctr,
		cancelFunc:    tokenAuthenticator.Close,
	}, nil
}
