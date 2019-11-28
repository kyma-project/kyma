package director

import (
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

type ApplicationsForRuntimeResponse struct {
	Result *ApplicationPage `json:"result"`
}

type SetRuntimeLabelResponse struct {
	Result *graphql.Label `json:"result"`
}

type ApplicationPage struct {
	Data       []*Application    `json:"data"`
	PageInfo   *graphql.PageInfo `json:"pageInfo"`
	TotalCount int               `json:"totalCount"`
}

type Application struct {
	ID          string                          `json:"id"`
	Name        string                          `json:"name"`
	Description *string                         `json:"description"`
	Labels      Labels                          `json:"labels"`
	APIs        *graphql.APIDefinitionPage      `json:"apis"`
	EventAPIs   *graphql.EventAPIDefinitionPage `json:"eventAPIs"`
	Documents   *graphql.DocumentPage           `json:"documents"`
	Auths       []*graphql.SystemAuth           `json:"auths"`
}

type Labels map[string][]string
