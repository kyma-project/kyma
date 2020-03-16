package main

import (
	"fmt"
	"time"

	"github.com/kyma-project/kyma/tests/compass-runtime-agent/test/testkit/authentication"

	"github.com/sirupsen/logrus"
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
	fatalOnError(err, "while reading configurations from environment variables")

	idProviderConfig := authentication.BuildIdProviderConfig(authentication.EnvConfig{
		Domain:        cfg.Domain,
		UserEmail:     cfg.TestUserEmail,
		UserPassword:  cfg.TestUserPassword,
		ClientTimeout: time.Second * 10,
	})

	token, err := authentication.GetToken(idProviderConfig)
	fatalOnError(err, "while geting token")

	return token
}

func fatalOnError(err error, context string) {
	if err != nil {
		logrus.Fatal(fmt.Sprintf("%s: %v", context, err))
	}
}
