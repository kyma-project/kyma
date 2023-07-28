package types

type CompassRuntimeAgentConfig struct {
	ConnectorUrl string
	RuntimeID    string
	Token        string
	Tenant       string
}

type RollbackFunc func() error

//go:generate mockery --name=DirectorClient
type DirectorClient interface {
	RegisterRuntime(runtimeName string) (string, error)
	RegisterFormation(formationName string) error
	AssignRuntimeToFormation(runtimeId, formationName string) error
	UnregisterRuntime(id string) error
	UnregisterFormation(formationName string) error
	GetConnectionToken(runtimeID string) (string, string, error)
}

//go:generate mockery --name=CompassConfigurator
type CompassConfigurator interface {
	Do(runtimeName, formationName string) (CompassRuntimeAgentConfig, RollbackFunc, error)
}

//go:generate mockery --name=DeploymentConfigurator
type DeploymentConfigurator interface {
	Do(caSecretName, clusterCertSecretName, runtimeAgentConfigSecretName, controllerSyncPeriodTime string) (RollbackFunc, error)
}

//go:generate mockery --name=CertificateSecretConfigurator
type CertificateSecretConfigurator interface {
	Do(caSecretName, clusterCertSecretName string) (RollbackFunc, error)
}

//go:generate mockery --name=ConfigurationSecretConfigurator
type ConfigurationSecretConfigurator interface {
	Do(configurationSecretName string, config CompassRuntimeAgentConfig) (RollbackFunc, error)
}

//go:generate mockery --name=CompassConnectionConfigurator
type CompassConnectionConfigurator interface {
	Do() (RollbackFunc, error)
}
