package director

import (
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

type ApplicationsAndLabelsForRuntimeResponse struct {
	Runtime          *Runtime         `json:"runtime"`
	ApplicationsPage *ApplicationPage `json:"applicationsForRuntime"`
}

type SetRuntimeLabelResponse struct {
	Result *graphql.Label `json:"setRuntimeLabel"`
}

type ApplicationPage struct {
	Data       []*Application    `json:"data"`
	PageInfo   *graphql.PageInfo `json:"pageInfo"`
	TotalCount int               `json:"totalCount"`
}

type Application struct {
	ID           string                  `json:"id"`
	Name         string                  `json:"name"`
	ProviderName *string                 `json:"providerName"`
	Description  *string                 `json:"description"`
	Labels       map[string]interface{}  `json:"labels"`
	Auths        []*graphql.SystemAuth   `json:"auths"`
	Packages     *graphql.PackagePageExt `json:"packages"`
}

type Runtime struct {
	Labels map[string]interface{} `json:"labels"`
}
