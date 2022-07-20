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
		RequestHeaders:         map[string][]string{"Hkey1": {"Hval1"}, "Hkey2": {"Hkey21", "Hkey22"}},
		RequestQueryParameters: map[string][]string{"Qkey1": {"Qval1"}, "Qkey2": {"Qkey21", "Qkey22"}},
	}
}

func (c *Config) String() string {
	return fmt.Sprintf("LogLevel: %s", c.LogLevel)
}
