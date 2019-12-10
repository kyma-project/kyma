package applications

import "github.com/kyma-incubator/compass/components/director/pkg/graphql"

type EventAPIDefinitionInput graphql.EventDefinitionInput

func NewEventAPI(name, description string) *EventAPIDefinitionInput {
	eventAPI := EventAPIDefinitionInput(graphql.EventDefinitionInput{
		Name:        name,
		Description: &description,
		Spec: &graphql.EventSpecInput{
			Data:   nil,
			Type:   graphql.EventSpecTypeAsyncAPI,
			Format: graphql.SpecFormatJSON,
		},
	})

	return &eventAPI
}

func (in *EventAPIDefinitionInput) WithJsonEventApiSpec(data *graphql.CLOB) *EventAPIDefinitionInput {
	in.Spec = &graphql.EventSpecInput{
		Data:   data,
		Type:   graphql.EventSpecTypeAsyncAPI,
		Format: graphql.SpecFormatJSON,
	}
	return in
}

func (in *EventAPIDefinitionInput) WithYamlEventApiSpec(data *graphql.CLOB) *EventAPIDefinitionInput {
	in.Spec = &graphql.EventSpecInput{
		Data:   data,
		Type:   graphql.EventSpecTypeAsyncAPI,
		Format: graphql.SpecFormatYaml,
	}
	return in
}

func (input *EventAPIDefinitionInput) ToCompassInput() *graphql.EventDefinitionInput {
	api := graphql.EventDefinitionInput(*input)
	return &api
}
