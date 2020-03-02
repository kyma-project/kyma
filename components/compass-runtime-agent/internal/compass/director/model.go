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
	ID               string                       `json:"id"`
	Name             string                       `json:"name"`
	ProviderName     *string                      `json:"providerName"`
	Description      *string                      `json:"description"`
	Labels           Labels                       `json:"labels"`
	APIDefinitions   *graphql.APIDefinitionPage   `json:"apiDefinitions"`
	EventDefinitions *graphql.EventDefinitionPage `json:"eventDefinitions"`
	Documents        *graphql.DocumentPage        `json:"documents"`
	Auths            []*graphql.SystemAuth        `json:"auths"`
	Packages         *graphql.PackagePage         `json:"packages"`
}

type Labels map[string]interface{}
