package busola

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/rest"

	"github.com/kyma-project/kyma/components/busola-migrator/internal/config"
)

func TestBuildInitURL(t *testing.T) {
	// GIVEN
	kubeConfig := rest.Config{Host: "example.com", TLSClientConfig: rest.TLSClientConfig{CAData: []byte{1, 0, 1, 0}}}
	appCfg := config.Config{
		BusolaURL: "https://busola.url",
		OIDC: config.OIDCConfig{
			IssuerURL: "https://account.url",
			ClientID:  "123",
			Scope:     "openid",
			UsePKCE:   false,
		},
	}
	urlRegexp := regexp.MustCompile(`^https://busola\.url/\?init=[0-9a-zA-Z-_]+$`)

	// WHEN
	res, err := BuildInitURL(appCfg, &kubeConfig)

	// THEN
	assert.NoError(t, err)
	assert.Regexp(t, urlRegexp, res)
}
