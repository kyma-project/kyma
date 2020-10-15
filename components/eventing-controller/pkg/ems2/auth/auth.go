package auth

import (
	httpclient2 "github.com/kyma-project/kyma/components/eventing-controller/pkg/ems2/httpclient"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
)

type Authenticator struct {
	client *httpclient2.Client
}

func NewAuthenticator() *Authenticator {
	authenticator := &Authenticator{}
	config := getDefaultOauth2Config(env.GetConfig())
	authenticator.client = httpclient2.NewHttpClient(config)
	return authenticator
}

func (a *Authenticator) GetClient() *httpclient2.Client {
	return a.client
}
