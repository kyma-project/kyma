package applications

import "github.com/kyma-incubator/compass/components/director/pkg/graphql"

type EventDefinitionInput graphql.EventDefinitionInput

func NewEventDefinition(name, description string) *EventDefinitionInput {
	eventAPI := EventDefinitionInput(graphql.EventDefinitionInput{
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

func (in *EventDefinitionInput) WithJsonEventSpec(data *graphql.CLOB) *EventDefinitionInput {
	in.Spec = &graphql.EventSpecInput{
		Data:   data,
		Type:   graphql.EventSpecTypeAsyncAPI,
		Format: graphql.SpecFormatJSON,
	}
	return in
}

func (in *EventDefinitionInput) WithYamlEventSpec(data *graphql.CLOB) *EventDefinitionInput {
	in.Spec = &graphql.EventSpecInput{
		Data:   data,
		Type:   graphql.EventSpecTypeAsyncAPI,
		Format: graphql.SpecFormatYaml,
	}
	return in
}

func (input *EventDefinitionInput) ToCompassInput() *graphql.EventDefinitionInput {
	api := graphql.EventDefinitionInput(*input)
	return &api
}
