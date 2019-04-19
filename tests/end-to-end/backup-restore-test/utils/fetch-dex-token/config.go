package fetch_dex_token

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

type Config struct {
	Domain       string `envconfig:"default=kyma.local"`
	UserEmail    string
	UserPassword string
}

type IdProviderConfig struct {
	DexConfig       DexConfig
	ClientConfig    ClientConfig
	UserCredentials UserCredentials
}

type DexConfig struct {
	BaseUrl           string
	AuthorizeEndpoint string
	TokenEndpoint     string
}

type ClientConfig struct {
	ID          string
	RedirectUri string
}

type UserCredentials struct {
	Username string
	Password string
}

func LoadConfig() (Config, error) {
	config := Config{}
	err := envconfig.Init(&config)
	if err != nil {
		return Config{}, errors.Wrap(err, "while loading environment variables")
	}

	return config, nil
}

func (c *Config) IdProviderConfig() IdProviderConfig {
	return IdProviderConfig{
		DexConfig: DexConfig{
			BaseUrl:           fmt.Sprintf("https://dex.%s", c.Domain),
			AuthorizeEndpoint: fmt.Sprintf("https://dex.%s/auth", c.Domain),
			TokenEndpoint:     fmt.Sprintf("https://dex.%s/token", c.Domain),
		},
		ClientConfig: ClientConfig{
			ID:          "kyma-client",
			RedirectUri: "http://127.0.0.1:5555/callback",
		},
		UserCredentials: UserCredentials{
			Username: c.UserEmail,
			Password: c.UserPassword,
		},
	}
}
