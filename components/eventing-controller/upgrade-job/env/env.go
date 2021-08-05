package env

// env package contains config variables

// Config for env variables
type Config struct {
	ReleaseName     		string `envconfig:"RELEASE" required:"true"`
	KymaNamespace    		string `envconfig:"KYMA_NAMESPACE" default:"kyma-system"`
	EventingControllerName  string `envconfig:"EVENTING_CONTROLLER_NAME" required:"true"`
	EventingPublisherName	string `envconfig:"EVENTING_PUBLISHER_NAME" required:"true"`
	LogFormat 				string `envconfig:"APP_LOG_FORMAT" default:"json"`
	LogLevel  				string `envconfig:"APP_LOG_LEVEL" default:"warn"`
}
