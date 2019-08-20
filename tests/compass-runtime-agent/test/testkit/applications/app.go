package applications

import "github.com/kyma-incubator/compass/components/director/pkg/graphql"

type ApplicationInput graphql.ApplicationInput

func NewApplication(name, description string, labels map[string][]string) *ApplicationInput {
	appLabels := graphql.Labels(labels)

	app := ApplicationInput(graphql.ApplicationInput{
		Name:        name,
		Description: &description,
		Labels:      &appLabels,
		Apis:        nil,
		EventAPIs:   nil,
		Documents:   nil,
	})

	return &app
}

func (input *ApplicationInput) ToCompassInput() graphql.ApplicationInput {
	return graphql.ApplicationInput(*input)
}

func (input *ApplicationInput) WithAPIs(apis []*APIDefinitionInput) *ApplicationInput {
	compassAPIs := make([]*graphql.APIDefinitionInput, len(apis))
	for i, api := range apis {
		compassAPIs[i] = api.ToCompassInput()
	}

	input.Apis = compassAPIs

	return input
}

func (input *ApplicationInput) WithEventAPIs(apis []*EventAPIDefinitionInput) *ApplicationInput {
	compassAPIs := make([]*graphql.EventAPIDefinitionInput, len(apis))
	for i, api := range apis {
		compassAPIs[i] = api.ToCompassInput()
	}

	input.EventAPIs = compassAPIs

	return input
}
