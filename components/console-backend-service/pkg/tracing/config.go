package tracing

type Config struct {
	Url         string `envconfig:"default=http://zipkin.kyma-system:9411/api/v1/spans"`
	Debug       bool   `envconfig:"default=true"`
	ServiceName string `envconfig:"default=console-backend-service"`
}
