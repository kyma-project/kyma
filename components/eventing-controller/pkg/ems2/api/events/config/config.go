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

func GetDefaultConfig() *Config {
	baseApiURL := "https://enterprise-messaging-pubsub.cfapps.sap.hana.ondemand.com/sap/cp-kernel/ems-ce/v1"
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
