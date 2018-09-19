package gqlschema

type RemoteEnvironment struct {
	Name                  string
	Description           string
	Services              []RemoteEnvironmentService
	enabledInEnvironments []string
}
