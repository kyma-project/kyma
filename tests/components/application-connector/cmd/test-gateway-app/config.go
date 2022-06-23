package main

import (
	"fmt"
)

type Config struct {
	LogLevel          string `envconfig:"default=info"`
	CaCertPath        string `envconfig:"default=/etc/secret-volume/ca.cert"`
	ServerCertPath    string `envconfig:"default=/etc/secret-volume/server.cert"`
	ServerKeyPath     string `envconfig:"default=/etc/secret-volume/server.key"`
	Port              int    `envconfig:"default=8080"`
	MtlsPort          int    `envconfig:"default=8090"`
	BasicAuthUser     string `envconfig:"default=user"`
	BasicAuthPassword string `envconfig:"default=passwd"`
	OAuthClientID     string `envconfig:"default=clientID"`
	OAuthClientSecret string `envconfig:"default=clientSecret"`
}

func (c *Config) String() string {
	return fmt.Sprintf("LogLevel: %s, CaCertPath: %s, ServerCertPath: %s", c.LogLevel, c.CaCertPath, c.ServerCertPath)
}
