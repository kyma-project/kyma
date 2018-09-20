package gqlschema

type RemoteEnvironment struct {
	Name                  string
	Description           string
	Labels                JSON
	Services              []RemoteEnvironmentService
	enabledInEnvironments []string
}
