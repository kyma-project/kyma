package webhookconfig

type Config struct {
	CfgMapName      string `envconfig:"default=webhook-configmap"`
	CfgMapNamespace string `envconfig:"default=kyma-system"`
}
