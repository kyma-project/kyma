package director

import (
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

type CreateRuntimeResponse struct {
	Result *graphql.Runtime `json:"result"`
}

type DeleteRuntimeResponse struct {
	Result *graphql.Runtime `json:"result"`
}

type CreateApplicationResponse struct {
	Result *graphql.Application `json:"result"`
}

type DeleteApplicationResponse struct {
	Result *graphql.Application `json:"result"`
}

type AssignFormationResponse struct {
	Result *graphql.Formation `json:"result"`
}

type OneTimeTokenResponse struct {
	Result *graphql.OneTimeTokenForRuntimeExt `json:"result"`
}

type ApplicationInput struct {
	Name        string          `json:"name"`
	Description *string         `json:"description"`
	Labels      *graphql.Labels `json:"labels"`
}
