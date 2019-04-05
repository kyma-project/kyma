package webhookconfig

type Config struct {
	WebhookCfgMapName              string        `envconfig:"default=webhook-config-map"`
	WebhookCfgMapNamespace         string        `envconfig:"default=kyma-system"`
}