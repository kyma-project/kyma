package gqlschema

type RemoteEnvironment struct {
	Name                  string
	Description           string
	Labels                Labels
	Services              []RemoteEnvironmentService
	enabledInEnvironments []string
}
