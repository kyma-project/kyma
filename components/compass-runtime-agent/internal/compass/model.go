package compass

import (
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

type ApplicationsForRuntimeResponse struct {
	Result *ApplicationPage `json:"result"`
}

type ApplicationPage struct {
	Data       []*Application    `json:"data"`
	PageInfo   *graphql.PageInfo `json:"pageInfo"`
	TotalCount int               `json:"totalCount"`
}

type Application struct {
	ID          string                     `json:"id"`
	Name        string                     `json:"name"`
	Description *string                    `json:"description"`
	Labels      Labels                     `json:"labels"`
	Status      *graphql.ApplicationStatus `json:"status"`
	//Webhooks       []*graphql.ApplicationWebhook    `json:"webhooks"`
	APIs           *graphql.APIDefinitionPage      `json:"apis"`
	EventAPIs      *graphql.EventAPIDefinitionPage `json:"eventAPIs"`
	Documents      *graphql.DocumentPage           `json:"documents"`
	HealthCheckURL *string                         `json:"healthCheckURL"`
}

type Labels map[string][]string
