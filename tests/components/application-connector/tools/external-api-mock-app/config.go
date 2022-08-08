package main

import (
	"fmt"
)

type mTLS struct {
	caCertPath     string
	serverCertPath string
	serverKeyPath  string
	port           int
}

type Config struct {
	LogLevel               string
	Port                   int
	BasicAuthUser          string
	BasicAuthPassword      string
	OAuthClientID          string
	OAuthClientSecret      string
	RequestHeaders         map[string][]string
	RequestQueryParameters map[string][]string
	mTLS                   mTLS
	mTLSExpiredCerts       mTLS
}

func NewConfig() *Config {
	return &Config{
		LogLevel:               "info",
		Port:                   8080,
		BasicAuthUser:          "user",
		BasicAuthPassword:      "passwd",
		OAuthClientID:          "clientID",
		OAuthClientSecret:      "clientSecret",
		RequestHeaders:         map[string][]string{"Hkey1": {"Hval1"}, "Hkey2": {"Hval21", "Hval22"}},
		RequestQueryParameters: map[string][]string{"Qkey1": {"Qval1"}, "Qkey2": {"Qval21", "Qval22"}},
		mTLS: mTLS{
			port:           8090,
			caCertPath:     "/etc/secret-volume/ca.crt",
			serverCertPath: "/etc/secret-volume/server.crt",
			serverKeyPath:  "/etc/secret-volume/server.key",
		},
		mTLSExpiredCerts: mTLS{
			port:           8091,
			caCertPath:     "/etc/expired-server-cert-volume/ca.crt",
			serverCertPath: "/etc/expired-server-cert-volume/server.crt",
			serverKeyPath:  "/etc/expired-server-cert-volume/server.key",
		},
	}
}

func (c *Config) String() string {
	return fmt.Sprintf("LogLevel: %s", c.LogLevel)
}
