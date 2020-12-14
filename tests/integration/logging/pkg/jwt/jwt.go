package jwt

import (
	"fmt"
	"time"

	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/authentication"
	"github.com/pkg/errors"

	"github.com/vrischmann/envconfig"
)

type config struct {
	Domain           string
	TestUserEmail    string
	TestUserPassword string
}

// GetToken retrieves jwt token from authentication package
func GetToken() (string, string, error) {
	var cfg config
	err := envconfig.Init(&cfg)
	if err != nil {
		return "", "", errors.Wrap(err, "cannot read configurations from environment variables")
	}

	idProviderConfig := authentication.BuildIdProviderConfig(authentication.EnvConfig{
		Domain:        cfg.Domain,
		UserEmail:     cfg.TestUserEmail,
		UserPassword:  cfg.TestUserPassword,
		ClientTimeout: time.Second * 10,
	})

	token, err := authentication.GetToken(idProviderConfig)
	if err != nil {
		return "", "", errors.Wrap(err, "cannot get JWT token")
	}
	return token, cfg.Domain, nil
}

// SetAuthHeader sets authorization header with JWT
func SetAuthHeader(token string) string {
	return fmt.Sprintf("Bearer %s", token)
}
