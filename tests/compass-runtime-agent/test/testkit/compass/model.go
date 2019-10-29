package compass

import (
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

type Application struct {
	ID          string                          `json:"id"`
	Name        string                          `json:"name"`
	Description *string                         `json:"description"`
	Labels      map[string][]string             `json:"labels"`
	APIs        *graphql.APIDefinitionPage      `json:"apis"`
	EventAPIs   *graphql.EventAPIDefinitionPage `json:"eventAPIs"`
	Documents   *graphql.DocumentPage           `json:"documents"`
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
