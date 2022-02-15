package oauth

import (
	"testing"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/env"
)

func TestConfig(t *testing.T) {
	t.Parallel()

	cfg := &env.BebConfig{ClientID: "someID", ClientSecret: "someSecret", TokenEndpoint: "someEndpoint"}
	conf := Config(cfg)

	if cfg.ClientID != conf.ClientID {
		t.Errorf("Client IDs do not match want:%s but got:%s", cfg.ClientID, conf.ClientID)
	}
	if cfg.ClientSecret != conf.ClientSecret {
		t.Errorf("Client secrets do not match want:%s but got:%s", cfg.ClientSecret, conf.ClientSecret)
	}
	if cfg.TokenEndpoint != conf.TokenURL {
		t.Errorf("Token URLs do not match want:%s but got:%s", cfg.TokenEndpoint, conf.TokenURL)
	}
}
