package graphql

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

type envConfig struct {
	Domain          string `envconfig:"default=kyma.local"`
	GraphQLEndpoint string `envconfig:"optional,GRAPHQL_ENDPOINT"`
	Username        string
	Password        string
	IsLocalCluster  bool `envconfig:"default=true"`
}

type config struct {
	GraphQlEndpoint   string
	LocalClusterHosts []string
	IdProviderConfig  idProviderConfig
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

	config := config{
		LocalClusterHosts: make([]string, 0, 2),
	}

	graphQLEndpoint := env.GraphQLEndpoint
	if graphQLEndpoint == "" {
		graphQLEndpoint = fmt.Sprintf("https://ui-api.%s/graphql", env.Domain)
	}

	config.GraphQlEndpoint = graphQLEndpoint

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

	if env.IsLocalCluster {
		config.addLocalClusterHost(config.GraphQlEndpoint)
		config.addLocalClusterHost(config.IdProviderConfig.DexConfig.BaseUrl)
	}

	return config, nil
}

func (c *config) addLocalClusterHost(host string) {
	url := strings.TrimPrefix(host, "http://")
	url = strings.TrimPrefix(host, "https://")
	url = strings.Split(url, "/")[0]

	c.LocalClusterHosts = append(c.LocalClusterHosts, url)
}
