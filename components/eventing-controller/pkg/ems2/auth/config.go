package auth

import (
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
)

type Config struct {
	ClientID      string
	ClientSecret  string
	TokenEndpoint string
}

func GetDefaultConfig() *Config {
	c := Config{}
	c.ClientID = env.GetConfig().ClientID
	c.ClientSecret = env.GetConfig().ClientSecret
	c.TokenEndpoint = env.GetConfig().TokenEndpoint
	return &c
}
