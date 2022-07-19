package main

import (
	"fmt"
)

type Config struct {
	LogLevel          string `envconfig:"default=info"`
	Port              int    `envconfig:"default=8080"`
	BasicAuthUser     string `envconfig:"default=user"`
	BasicAuthPassword string `envconfig:"default=passwd"`
	OAuthClientID     string `envconfig:"default=clientID"`
	OAuthClientSecret string `envconfig:"default=clientSecret"`
	//TODO:
	RequestHeaders         map[string][]string `envconfig:"-"`
	RequestQueryParameters map[string][]string `envconfig:"-"`
}

func (c *Config) String() string {
	return fmt.Sprintf("LogLevel: %s", c.LogLevel)
}
