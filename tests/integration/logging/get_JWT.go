package main

import (
	"log"
	"time"

	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/authentication"

	"github.com/vrischmann/envconfig"
)

type config struct {
	Domain           string
	TestUserEmail    string
	TestUserPassword string
}

func getJWT() string {
	var cfg config
	err := envconfig.Init(&cfg)
	if err != nil {
		log.Fatalf("Error while reading configurations from environment variables: %v", err)
	}

	idProviderConfig := authentication.BuildIdProviderConfig(authentication.EnvConfig{
		Domain:        cfg.Domain,
		UserEmail:     cfg.TestUserEmail,
		UserPassword:  cfg.TestUserPassword,
		ClientTimeout: time.Second * 10,
	})

	token, err := authentication.GetToken(idProviderConfig)
	if err != nil {
		log.Fatalf("Error while while getting JWT token: %v", err)
	}
	return token
}
