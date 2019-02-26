package loader

type Config struct {
	TemporaryDirectory string `envconfig:"default=/tmp"`
}
