package tracing

type Config struct {
	CollectorUrl    string `envconfig:"default=http://zipkin.kyma-system:9411/api/v1/spans"`
	Debug           bool   `envconfig:"default=false"`
	ServiceSpanName string `envconfig:"default=console-backend-service"`
}
