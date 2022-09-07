package director

import (
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

type CreateApplicationResponse struct {
	Result *graphql.Application `json:"result"`
}

type DeleteApplicationResponse struct {
	Result *graphql.Application `json:"result"`
}

// not sure if this will be needed
type OneTimeTokenResponse struct {
	Result *graphql.OneTimeTokenForApplicationExt `json:"result"`
}

type ApplicationInput struct {
	Name        string          `json:"name"`
	Description *string         `json:"description"`
	Labels      *graphql.Labels `json:"labels"`
}
