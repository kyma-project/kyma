package auth

import (
	"net/http"

	"golang.org/x/oauth2"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/signals"
)

func NewAuthenticatedClient(cfg env.Config) *http.Client {
	ctx := signals.NewReusableContext()
	config := getDefaultOauth2Config(cfg)
	// create and configure oauth2 client
	client := config.Client(ctx)

	var base = http.DefaultTransport.(*http.Transport).Clone()
	client.Transport.(*oauth2.Transport).Base = base

	// TODO: Support tracing in eventing-controller #9767: https://github.com/kyma-project/kyma/issues/9767
	return client
}
