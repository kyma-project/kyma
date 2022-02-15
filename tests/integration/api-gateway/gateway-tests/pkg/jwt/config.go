package jwt

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

type envConfig struct {
	Domain        string        `envconfig:"TEST_DOMAIN,default=kyma.local"`
	UserEmail     string        `envconfig:"TEST_USER_EMAIL,default=admin@kyma.cx"`
	UserPassword  string        `envconfig:"TEST_USER_PASSWORD,default=1234"`
	ClientTimeout time.Duration `envconfig:"TEST_CLIENT_TIMEOUT,default=10s"` //Don't forget the unit!
}

//Config JWT configuration structure
type Config struct {
	OidcHydraConfig oidcHydraConfig
	EnvConfig       envConfig
}

type oidcHydraConfig struct {
	HydraFqdn       string
	ClientConfig    clientConfig
	UserCredentials userCredentials
}

type clientConfig struct {
	ID             string
	RedirectUri    string
	TimeoutSeconds time.Duration
}

type userCredentials struct {
	Username string
	Password string
}

//LoadConfig Generate test config from envs
func LoadConfig(oauthClientID string) (Config, error) {
	env := envConfig{}
	err := envconfig.Init(&env)
	if err != nil {
		return Config{}, errors.Wrap(err, "while loading environment variables")
	}

	config := Config{EnvConfig: env}
	config.OidcHydraConfig = oidcHydraConfig{
		HydraFqdn: fmt.Sprintf("oauth2.%s", env.Domain),
		ClientConfig: clientConfig{
			ID:             oauthClientID,
			RedirectUri:    "http://testclient3.example.com",
			TimeoutSeconds: env.ClientTimeout,
		},
		UserCredentials: userCredentials{
			Username: env.UserEmail,
			Password: env.UserPassword,
		},
	}

	return config, nil
}
