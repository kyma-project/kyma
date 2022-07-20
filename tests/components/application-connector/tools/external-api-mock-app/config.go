package main

import (
	"fmt"
)

type Config struct {
	LogLevel               string
	Port                   int
	BasicAuthUser          string
	BasicAuthPassword      string
	OAuthClientID          string
	OAuthClientSecret      string
	RequestHeaders         map[string][]string
	RequestQueryParameters map[string][]string
}

func NewConfig() *Config {
	return &Config{
		LogLevel:               "info",
		Port:                   8080,
		BasicAuthUser:          "user",
		BasicAuthPassword:      "passwd",
		OAuthClientID:          "clientID",
		OAuthClientSecret:      "clientSecret",
		RequestHeaders:         map[string][]string{"hkey1": {"hval1"}, "hkey2": {"hkey21", "hkey22"}},
		RequestQueryParameters: map[string][]string{"qkey1": {"qval1"}, "qkey2": {"qkey21", "qkey22"}},
	}
}

func (c *Config) String() string {
	return fmt.Sprintf("LogLevel: %s", c.LogLevel)
}
