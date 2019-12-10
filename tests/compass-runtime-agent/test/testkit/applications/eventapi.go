package applications

import "github.com/kyma-incubator/compass/components/director/pkg/graphql"

type EventAPIDefinitionInput graphql.EventAPIDefinitionInput

func NewEventAPI(name, description string) *EventAPIDefinitionInput {
	eventAPI := EventAPIDefinitionInput(graphql.EventAPIDefinitionInput{
		Name:        name,
		Description: &description,
		Spec: &graphql.EventAPISpecInput{
			Data:          nil,
			EventSpecType: graphql.EventAPISpecTypeAsyncAPI,
			Format:        graphql.SpecFormatJSON,
		},
	})

	return &eventAPI
}

func (in *EventAPIDefinitionInput) WithJsonEventApiSpec(data *graphql.CLOB) *EventAPIDefinitionInput {
	in.Spec = &graphql.EventAPISpecInput{
		Data:          data,
		EventSpecType: graphql.EventAPISpecTypeAsyncAPI,
		Format:        graphql.SpecFormatJSON,
	}
	return in
}

func (in *EventAPIDefinitionInput) WithYamlEventApiSpec(data *graphql.CLOB) *EventAPIDefinitionInput {
	in.Spec = &graphql.EventAPISpecInput{
		Data:          data,
		EventSpecType: graphql.EventAPISpecTypeAsyncAPI,
		Format:        graphql.SpecFormatYaml,
	}
	return in
}

func (in *EventAPIDefinitionInput) WithXMLEventApiSpec(data *graphql.CLOB) *EventAPIDefinitionInput {
	in.Spec = &graphql.EventAPISpecInput{
		Data:          data,
		EventSpecType: graphql.EventAPISpecTypeAsyncAPI,
		Format:        graphql.SpecFormatXML,
	}
	return in
}

func (input *EventAPIDefinitionInput) ToCompassInput() *graphql.EventAPIDefinitionInput {
	api := graphql.EventAPIDefinitionInput(*input)
	return &api
}
