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
	IdProviderConfig idProviderConfig
	EnvConfig        envConfig
}

type idProviderConfig struct {
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

	config.IdProviderConfig = idProviderConfig{
		DexConfig: dexConfig{
			BaseUrl:           fmt.Sprintf("https://dex.%s", env.Domain),
			AuthorizeEndpoint: fmt.Sprintf("https://dex.%s/auth", env.Domain),
			TokenEndpoint:     fmt.Sprintf("https://dex.%s/token", env.Domain),
		},
		ClientConfig: clientConfig{
			ID:          "kyma-client",
			RedirectUri: "http://127.0.0.1:5555/callback",
		},
	}

	config.IdProviderConfig.UserCredentials = userCredentials{
		Username: env.UserEmail,
		Password: env.UserPassword,
	}

	return config, nil
}
