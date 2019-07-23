package compass

import (
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

type Application struct {
	ID             string                       `json:"id"`
	Name           string                       `json:"name"`
	Description    *string                      `json:"description"`
	Labels         Labels                       `json:"labels"`
	Status         *graphql.ApplicationStatus   `json:"status"`
	Webhooks       *graphql.Webhook             `json:"webhooks"`
	APIs           []graphql.APIDefinition      `json:"apis"`
	EventAPIs      []graphql.EventAPIDefinition `json:"eventAPIs"`
	Documents      []graphql.Document           `json:"documents"`
	HealthCheckURL *string                      `json:"healthCheckURL"`
}

type Labels map[string][]string
