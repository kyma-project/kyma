package main

import (
	"fmt"
)

type Config struct {
	LogLevel          string
	Port              int
	MtlsPort          int
	BasicAuthUser     string
	BasicAuthPassword string
	OAuthClientID     string
	OAuthClientSecret string
	RequestHeaders         map[string][]string
	RequestQueryParameters map[string][]string
	CaCertPath        string
	ServerCertPath    string
	ServerKeyPath     string

}

func NewConfig() *Config {
	return &Config{
		LogLevel:               "info",
		Port:                   8080,
		MtlsPort: 8090,
		BasicAuthUser:          "user",
		BasicAuthPassword:      "passwd",
		OAuthClientID:          "clientID",
		OAuthClientSecret:      "clientSecret",
		RequestHeaders:         map[string][]string{"Hkey1": {"Hval1"}, "Hkey2": {"Hval21", "Hval22"}},
		RequestQueryParameters: map[string][]string{"Qkey1": {"Qval1"}, "Qkey2": {"Qval21", "Qval22"}},
		CaCertPath: "",
		ServerCertPath: "",
		ServerKeyPath: "",
	}
}

func (c *Config) String() string {
	return fmt.Sprintf("LogLevel: %s", c.LogLevel)
}
