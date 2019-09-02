package loader

type Config struct {
	TemporaryDirectory string `envconfig:"default=/tmp"`
	VerifySSL          bool   `envconfig:"default=true"`
}
