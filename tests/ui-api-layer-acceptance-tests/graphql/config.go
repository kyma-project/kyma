package graphql

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

type envConfig struct {
	Domain          string `envconfig:"default=kyma.local"`
	GraphQLEndpoint string `envconfig:"optional,GRAPHQL_ENDPOINT"`
	Username        string
	Password        string
}

type config struct {
	GraphQLEndpoint  string
	IdProviderConfig idProviderConfig
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

func loadConfig() (config, error) {
	env := envConfig{}
	err := envconfig.Init(&env)
	if err != nil {
		return config{}, errors.Wrap(err, "while loading environment variables")
	}

	config := config{}

	graphQLEndpoint := env.GraphQLEndpoint
	if graphQLEndpoint == "" {
		graphQLEndpoint = fmt.Sprintf("https://ui-api.%s/graphql", env.Domain)
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

		UserCredentials: userCredentials{
			Username: env.Username,
			Password: env.Password,
		},
	}

	return config, nil
}
