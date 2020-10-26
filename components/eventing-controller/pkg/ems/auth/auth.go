package auth

import (
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/httpclient"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
)

type Authenticator struct {
	client *httpclient.Client
}

func NewAuthenticator() *Authenticator {
	authenticator := &Authenticator{}
	config := getDefaultOauth2Config(env.GetConfig())
	authenticator.client = httpclient.NewHttpClient(config)
	return authenticator
}

func (a *Authenticator) GetClient() *httpclient.Client {
	return a.client
}
