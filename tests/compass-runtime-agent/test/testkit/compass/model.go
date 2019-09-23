package compass

import "github.com/kyma-incubator/compass/components/director/pkg/graphql"

type Application struct {
	ID          string                          `json:"id"`
	Name        string                          `json:"name"`
	Description *string                         `json:"description"`
	Labels      map[string][]string             `json:"labels"`
	APIs        *graphql.APIDefinitionPage      `json:"apis"`
	EventAPIs   *graphql.EventAPIDefinitionPage `json:"eventAPIs"`
	Documents   *graphql.DocumentPage           `json:"documents"`
}

type ApplicationResponse struct {
	Result Application `json:"result"`
}

type APIResponse struct {
	Result *graphql.APIDefinition `json:"result"`
}

type CreateEventAPIResponse struct {
	Result *graphql.EventAPIDefinition `json:"result"`
}

type DeleteResponse struct {
	Result struct {
		ID string `json:"id"`
	} `json:"result"`
}

type IdResponse struct {
	Id string `json:"id"`
}

type SetRuntimeLabelResponse struct {
	Result struct {
		Key   string      `json:"key"`
		Value interface{} `json:"value"`
	} `json:"result"`
}

type ScenarioLabelDefinitionResponse struct {
	Result struct {
		Key    string              `json:"key"`
		Schema *graphql.JSONSchema `json:"schema"`
	} `json:"result"`
}
