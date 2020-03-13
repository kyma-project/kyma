package gqlschema

type Application struct {
	Name                string
	Description         string
	Labels              Labels
	Services            []ApplicationService
	enabledInNamespaces []string
	CompassMetadata     CompassMetadata
}
