package auth

import (
	"log"

	"github.com/vrischmann/envconfig"
)

type Config struct {
	ClientID      string
	ClientSecret  string
	TokenEndpoint string
	// TODO for OAuth2 secured webhooks
	//SubscriptionClientID string
	//SubscriptionClientSecret string
	//SubscriptionTokenUrl string
}

func GetDefaultConfig() *Config {
	c := Config{}
	if err := envconfig.InitWithPrefix(&c, "EMS"); err != nil {
		log.Fatalf("Did not find required EMS environment variables: %v", err)
	}

	return &c
}
