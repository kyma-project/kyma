package main

import (
	"fmt"
)

type Config struct {
	LogLevel          string `envconfig:"default=info"`
	MTLSCertPath      string `envconfig:"default=/etc/secret-volume/"`
	Port              int    `envconfig:"default=8080"`
	MtlsPort          int    `envconfig:"default=8090"`
	BasicAuthUser     string `envconfig:"default=user"`
	BasicAuthPassword string `envconfig:"default=passwd"`
	OAuthClientID     string `envconfig:"default=clientID"`
	OAuthClientSecret string `envconfig:"default=clientSecret"`
}

func (c *Config) String() string {
	return fmt.Sprintf("LogLevel: %s, MTLSCertPath: %s" + c.LogLevel + c.MTLSCertPath)
}
