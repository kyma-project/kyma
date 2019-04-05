package webhookconfig

type Config struct {
	WebhookCfgMapName      string `envconfig:"default=webhook-configmap"`
	WebhookCfgMapNamespace string `envconfig:"default=kyma-system"`
}
