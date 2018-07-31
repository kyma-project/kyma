package graphql

import (
	"fmt"
	"os"
	"strings"
)

type config struct {
	graphQlEndpoint   string
	localClusterHosts []string
	idProviderConfig  idProviderConfig
}

type idProviderConfig struct {
	dexConfig       dexConfig
	clientConfig    clientConfig
	userCredentials userCredentials
}

type dexConfig struct {
	baseUrl           string
	authorizeEndpoint string
	tokenEndpoint     string
}

type clientConfig struct {
	id          string
	redirectUri string
}

type userCredentials struct {
	username string
	password string
}

func loadConfig() config {

	config := config{
		localClusterHosts: make([]string, 0, 2),
	}

	domain := os.Getenv("DOMAIN")
	isLocalClusterStr := strings.ToLower(os.Getenv("IS_LOCAL_CLUSTER"))
	graphQlEndpoint := os.Getenv("GRAPHQL_ENDPOINT")

	if domain == "" {
		domain = "kyma.local"
	}
	isLocalCluster := isLocalClusterStr == "" || isLocalClusterStr == "true" || isLocalClusterStr == "yes" || isLocalClusterStr == "y"

	if graphQlEndpoint == "" {

		config.graphQlEndpoint = fmt.Sprintf("https://ui-api.%s/graphql", domain)
		if isLocalCluster {
			config.addLocalClusterHost(fmt.Sprintf("ui-api.%s", domain))
		}
	}

	config.idProviderConfig = idProviderConfig{
		dexConfig: dexConfig{
			baseUrl:           fmt.Sprintf("https://dex.%s", domain),
			authorizeEndpoint: fmt.Sprintf("https://dex.%s/auth", domain),
			tokenEndpoint:     fmt.Sprintf("https://dex.%s/token", domain),
		},
		clientConfig: clientConfig{
			id:          "kyma-client",
			redirectUri: "http://127.0.0.1:5555/callback",
		},

		userCredentials: userCredentials{
			username: "admin@kyma.cx",
			password: "nimda123",
		},
	}

	if isLocalCluster {
		config.addLocalClusterHost(fmt.Sprintf("dex.%s", domain))
	}

	return config
}

func (c *config) addLocalClusterHost(host string) {
	c.localClusterHosts = append(c.localClusterHosts, host)
}
