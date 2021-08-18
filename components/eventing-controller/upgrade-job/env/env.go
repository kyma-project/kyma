package env

// env package contains config variables

// Config for env variables
type Config struct {
	ReleaseName            string `envconfig:"RELEASE" required:"true"`
	Domain                 string `envconfig:"DOMAIN" required:"true"` // Domain holds the Kyma domain
	KymaNamespace          string `envconfig:"KYMA_NAMESPACE" default:"kyma-system"`
	EventingControllerName string `envconfig:"EVENTING_CONTROLLER_NAME" default:"eventing-controller"`
	LogFormat              string `envconfig:"APP_LOG_FORMAT" default:"json"`
	LogLevel               string `envconfig:"APP_LOG_LEVEL" default:"warn"`
}
