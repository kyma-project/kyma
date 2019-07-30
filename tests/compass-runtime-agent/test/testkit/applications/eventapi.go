package applications

import "github.com/kyma-incubator/compass/components/director/pkg/graphql"

type EventAPIDefinitionInput graphql.EventAPIDefinitionInput

func NewEventAPI(name, description string) EventAPIDefinitionInput {
	return EventAPIDefinitionInput(graphql.EventAPIDefinitionInput{
		Name:        name,
		Description: &description,
		Spec:        nil, // TODO  - test
	})
}

func (input *EventAPIDefinitionInput) ToCompassInput() *graphql.EventAPIDefinitionInput {
	api := graphql.EventAPIDefinitionInput(*input)
	return &api
}
