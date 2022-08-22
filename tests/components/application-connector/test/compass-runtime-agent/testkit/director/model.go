package director

import (
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

type CreateApplicationResponse struct {
	Result *graphql.Application `json:"result"`
}

//type GetRuntimeResponse struct {
//	Result *graphql.RuntimeExt `json:"result"`
//}
//
//type GetRuntimesResponse struct {
//	Result *graphql.RuntimePage `json:"result"`
//}
//
//type DeleteApplicationResponse struct {
//	Result *graphql.Runtime `json:"result"`
//}

type OneTimeTokenResponse struct {
	Result *graphql.OneTimeTokenForApplicationExt `json:"result"`
}

type ApplicationInput struct {
	Name        string          `json:"name"`
	Description *string         `json:"description"`
	Labels      *graphql.Labels `json:"labels"`
}
