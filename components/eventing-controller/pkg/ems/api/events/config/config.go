package config

import "github.com/kyma-project/kyma/components/eventing-controller/pkg/env"

type Config struct {
	BaseURL              string
	PublishURL           string
	CreateURL            string
	ListURL              string
	GetURLFormat         string
	DeleteURLFormat      string
	HandshakeURLFormat   string
	UpdateStateURLFormat string
}

func GetDefaultConfig() *Config {
	baseApiURL := env.GetConfig().BebApiUrl
	return &Config{
		BaseURL:              baseApiURL,
		PublishURL:           baseApiURL + "/events",
		CreateURL:            baseApiURL + "/events/subscriptions",
		ListURL:              baseApiURL + "/events/subscriptions",
		GetURLFormat:         baseApiURL + "/events/subscriptions/%s",
		DeleteURLFormat:      baseApiURL + "/events/subscriptions/%s",
		HandshakeURLFormat:   baseApiURL + "/events/subscriptions/%s/handshake",
		UpdateStateURLFormat: baseApiURL + "/events/subscriptions/%s/state",
	}
}
