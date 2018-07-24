package gqlschema

type RemoteEnvironmentService struct {
	ID                  string
	DisplayName         string
	LongDescription     string
	ProviderDisplayName string
	Tags                []string
	Entries             []RemoteEnvironmentEntry
}
