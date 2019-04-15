package fetch_dex_token

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

type envConfig struct {
	Domain       string `envconfig:"default=kyma.local"`
	UserEmail    string
	UserPassword string
}

type Config struct {
	EnvConfig        envConfig
}

type IdProviderConfig struct {
	DexConfig       dexConfig
	ClientConfig    clientConfig
	UserCredentials userCredentials
}

type dexConfig struct {
	BaseUrl           string
	AuthorizeEndpoint string
	TokenEndpoint     string
}

type clientConfig struct {
	ID          string
	RedirectUri string
}

type userCredentials struct {
	Username string
	Password string
}

func LoadConfig() (Config, error) {
	env := envConfig{}
	err := envconfig.Init(&env)
	if err != nil {
		return Config{}, errors.Wrap(err, "while loading environment variables")
	}

	config := Config{EnvConfig: env}

	return config, nil
}

func (c *Config) IdProviderConfig() IdProviderConfig {
	return IdProviderConfig{
		DexConfig: dexConfig{
			BaseUrl:           fmt.Sprintf("https://dex.%s", c.EnvConfig.Domain),
			AuthorizeEndpoint: fmt.Sprintf("https://dex.%s/auth", c.EnvConfig.Domain),
			TokenEndpoint:     fmt.Sprintf("https://dex.%s/token", c.EnvConfig.Domain),
		},
		ClientConfig: clientConfig{
			ID:          "kyma-client",
			RedirectUri: "http://127.0.0.1:5555/callback",
		},
		UserCredentials: userCredentials{
			Username: c.EnvConfig.UserEmail,
			Password: c.EnvConfig.UserPassword,
		},
	}
}
