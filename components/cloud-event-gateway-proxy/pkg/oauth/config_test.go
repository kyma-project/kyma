package oauth

import (
	"testing"

	"github.com/kyma-project/kyma/components/cloud-event-gateway-proxy/pkg/gateway"
)

func TestConfig(t *testing.T) {
	t.Parallel()

	env := gateway.EnvConfig{ClientID: "someID", ClientSecret: "someSecret", TokenEndpoint: "someEndpoint"}
	conf := Config(env)

	if env.ClientID != conf.ClientID {
		t.Errorf("Client IDs do not match want:%s but got:%s", env.ClientID, conf.ClientID)
	}
	if env.ClientSecret != conf.ClientSecret {
		t.Errorf("Client secrets do not match want:%s but got:%s", env.ClientSecret, conf.ClientSecret)
	}
	if env.TokenEndpoint != conf.TokenURL {
		t.Errorf("Token URLs do not match want:%s but got:%s", env.TokenEndpoint, conf.TokenURL)
	}
}
