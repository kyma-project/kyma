package config

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

func GetDefaultConfig(baseAPIURL string) *Config {
	return &Config{
		BaseURL:              baseAPIURL,
		PublishURL:           baseAPIURL + "/events",
		CreateURL:            baseAPIURL + "/events/subscriptions",
		ListURL:              baseAPIURL + "/events/subscriptions",
		GetURLFormat:         baseAPIURL + "/events/subscriptions/%s",
		DeleteURLFormat:      baseAPIURL + "/events/subscriptions/%s",
		HandshakeURLFormat:   baseAPIURL + "/events/subscriptions/%s/handshake",
		UpdateStateURLFormat: baseAPIURL + "/events/subscriptions/%s/state",
	}
}
