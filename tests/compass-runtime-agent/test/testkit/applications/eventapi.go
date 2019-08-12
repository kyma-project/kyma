package applications

import "github.com/kyma-incubator/compass/components/director/pkg/graphql"

type EventAPIDefinitionInput graphql.EventAPIDefinitionInput

func NewEventAPI(name, description string) *EventAPIDefinitionInput {
	eventAPI := EventAPIDefinitionInput(graphql.EventAPIDefinitionInput{
		Name:        name,
		Description: &description,
		Spec: &graphql.EventAPISpecInput{
			Data:          nil, // TODO: Allow to pass spec when Asset Store is ready
			EventSpecType: "ASYNC_API",
			Format:        "JSON",
		},
	})

	return &eventAPI
}

func (input *EventAPIDefinitionInput) ToCompassInput() *graphql.EventAPIDefinitionInput {
	api := graphql.EventAPIDefinitionInput(*input)
	return &api
}
