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
	Domain               string `envconfig:"default=kyma.local"`
	GraphQLEndpoint      string `envconfig:"optional,GRAPHQL_ENDPOINT"`
	AdminEmail           string
	AdminPassword        string
	ReadOnlyUserEmail    string
	ReadOnlyUserPassword string
	NoRightsUserEmail    string
	NoRightsUserPassword string
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

func loadConfig(userContext User) (config, error) {
	env := envConfig{}
	err := envconfig.Init(&env)
	if err != nil {
		return config{}, errors.Wrap(err, "while loading environment variables")
	}

	config := config{EnvConfig: env}

	graphQLEndpoint := env.GraphQLEndpoint
	if graphQLEndpoint == "" {
		graphQLEndpoint = fmt.Sprintf("https://console-backend.%s/graphql", env.Domain)
	}
	config.GraphQLEndpoint = graphQLEndpoint

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

	switch userContext {

	case AdminUser:
		config.IdProviderConfig.UserCredentials = userCredentials{
			Username: env.AdminEmail,
			Password: env.AdminPassword,
		}
	case ReadOnlyUser:
		config.IdProviderConfig.UserCredentials = userCredentials{
			Username: env.ReadOnlyUserEmail,
			Password: env.ReadOnlyUserPassword,
		}
	case NoRightsUser:
		config.IdProviderConfig.UserCredentials = userCredentials{
			Username: env.NoRightsUserEmail,
			Password: env.NoRightsUserPassword,
		}
	}

	return config, nil
}
