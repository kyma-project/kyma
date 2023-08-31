package webhook

type Config struct {
	SystemNamespace string `envconfig:"default=kyma-system"`
	ServiceName     string `envconfig:"default=serverless-webhook"`
	SecretName      string `envconfig:"default=serverless-webhook"`
	Port            int    `envconfig:"default=8443"`
	LogConfigPath   string `envconfig:"default=/appdata/log_config.yaml"`
	ConfigPath      string `envconfig:"default=/appdata/config.yaml"`
}
