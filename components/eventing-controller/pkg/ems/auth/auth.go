package auth

import (
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/httpclient"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
)

type Authenticator struct {
	client *httpclient.Client
}

func NewAuthenticatedClient(cfg env.Config) *http.Client {
	authenticator := &Authenticator{}
	config := getDefaultOauth2Config(cfg)
	// create and configure oauth2 client
	client := cfg.Client(ctx)

	var base = http.DefaultTransport.(*http.Transport).Clone()
	client.Transport.(*oauth2.Transport).Base = base

	// TODO: Support tracing in eventing-controller #9767: https://github.com/kyma-project/kyma/issues/9767
	return client

}

// NewClient returns a new HTTP client which have nested transports for handling oauth2 security, HTTP connection pooling, and tracing.
func newOauth2Client(ctx context.Context, cfg clientcredentials.Config) *http.Client {
}
