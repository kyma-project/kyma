package graphql

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

type User string

const (
	AdminUser    User = "admin"
	ReadOnlyUser User = "read-only"
	NoRightsUser User = "no-rights"
	NoUser       User = "no-user"
)

type envConfig struct {
	Domain        string `envconfig:"default=kyma.local"`
	AdminEmail    string
	AdminPassword string
}

type config struct {
	GraphQLEndpoint  string
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

func LoadConfig(userContext User) (config, error) {
	env := envConfig{}
	err := envconfig.Init(&env)
	if err != nil {
		return config{}, errors.Wrap(err, "while loading environment variables")
	}

	config := config{EnvConfig: env}

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
		Username: env.AdminEmail,
		Password: env.AdminPassword,
	}

	return config, nil
}
