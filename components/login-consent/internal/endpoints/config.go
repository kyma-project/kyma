package endpoints

import (
	hydraAPI "github.com/ory/hydra-client-go/models"
)

type HydraLoginConsentClient interface {
	GetLoginRequest(challenge string) (*hydraAPI.LoginRequest, error)
	AcceptLoginRequest(challenge string, body *hydraAPI.AcceptLoginRequest) (*hydraAPI.CompletedRequest, error)
	RejectLoginRequest(challenge string, body *hydraAPI.RejectRequest) (*hydraAPI.CompletedRequest, error)
	GetConsentRequest(challenge string) (*hydraAPI.ConsentRequest, error)
	AcceptConsentRequest(challenge string, body *hydraAPI.AcceptConsentRequest) (*hydraAPI.CompletedRequest, error)
}

type Config struct {
	hydraClient   HydraLoginConsentClient
	authenticator *Authenticator
}

func New(hydraClient HydraLoginConsentClient, authn *Authenticator) (*Config, error) {

	return &Config{
		hydraClient:   hydraClient,
		authenticator: authn,
	}, nil
}
