package env

// env package contains config variables

// Config for env variables
type Config struct {
	ReleaseName            string `envconfig:"RELEASE" required:"true"`
	KymaNamespace          string `envconfig:"KYMA_NAMESPACE" default:"kyma-system"`
	EventingControllerName string `envconfig:"EVENTING_CONTROLLER_NAME" default:"eventing-controller"`
	EventingPublisherName  string `envconfig:"EVENTING_PUBLISHER_NAME" default:"eventing-publisher-proxy"`
	LogFormat              string `envconfig:"APP_LOG_FORMAT" default:"json"`
	LogLevel               string `envconfig:"APP_LOG_LEVEL" default:"warn"`
}
