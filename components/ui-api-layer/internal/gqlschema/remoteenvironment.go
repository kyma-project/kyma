package gqlschema

type RemoteEnvironment struct {
	Name                  string
	Description           string
	Source                RemoteEnvironmentSource
	Services              []RemoteEnvironmentService
	enabledInEnvironments []string
}
