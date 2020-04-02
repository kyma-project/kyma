package compass

import (
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

type Application struct {
	ID          string                  `json:"id"`
	Name        string                  `json:"name"`
	Description *string                 `json:"description"`
	Labels      map[string]interface{}  `json:"labels"`
	Packages    *graphql.PackagePageExt `json:"packages"`
}

// GetContext is a helper function that returns Application ID and Name in well formatted string (for logging)
func (a Application) GetContext() string {
	return fmt.Sprintf("Application ID: %s, Application Name: %s", a.ID, a.Name)
}

type Runtime struct {
	graphql.Runtime
	Labels graphql.Labels `json:"labels"`
}

type graphQLResponseWrapper struct {
	Result interface{} `json:"result"`
}

type IdResponse struct {
	Id string `json:"id"`
}

type ScenarioLabelDefinition struct {
	Key    string              `json:"key"`
	Schema *graphql.JSONSchema `json:"schema"`
}
