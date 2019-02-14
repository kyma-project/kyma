package store

type Config struct {
	Endpoint         string `envconfig:"default=minio.kyma.local"`
	ExternalEndpoint string `envconfig:"default=https://minio.kyma.local"`
	AccessKey        string `envconfig:""`
	SecretKey        string `envconfig:""`
	UseSSL           bool   `envconfig:"default=true"`
}
