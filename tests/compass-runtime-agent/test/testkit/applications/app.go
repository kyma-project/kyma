package applications

import "github.com/kyma-incubator/compass/components/director/pkg/graphql"

type ApplicationRegisterInput graphql.ApplicationRegisterInput

func NewApplication(name, providerName, description string, labels map[string]interface{}) *ApplicationRegisterInput {
	appLabels := graphql.Labels(labels)

	app := ApplicationRegisterInput(graphql.ApplicationRegisterInput{
		Name:             name,
		ProviderName:     providerName,
		Description:      &description,
		Labels:           &appLabels,
		APIDefinitions:   nil,
		EventDefinitions: nil,
		Documents:        nil,
	})

	return &app
}

func (input *ApplicationRegisterInput) ToCompassInput() graphql.ApplicationRegisterInput {
	return graphql.ApplicationRegisterInput(*input)
}

func (input *ApplicationRegisterInput) WithAPIDefinitions(apis []*APIDefinitionInput) *ApplicationRegisterInput {
	compassAPIs := make([]*graphql.APIDefinitionInput, len(apis))
	for i, api := range apis {
		compassAPIs[i] = api.ToCompassInput()
	}

	input.APIDefinitions = compassAPIs

	return input
}

func (input *ApplicationRegisterInput) WithEventDefinitions(apis []*EventDefinitionInput) *ApplicationRegisterInput {
	compassAPIs := make([]*graphql.EventDefinitionInput, len(apis))
	for i, api := range apis {
		compassAPIs[i] = api.ToCompassInput()
	}

	input.EventDefinitions = compassAPIs

	return input
}

type ApplicationUpdateInput graphql.ApplicationUpdateInput

func NewApplicationUpdateInput(name, providerName, description string) *ApplicationUpdateInput {
	appUpdateInput := ApplicationUpdateInput(graphql.ApplicationUpdateInput{
		Name:         name,
		ProviderName: providerName,
		Description:  &description,
	})

	return &appUpdateInput
}

func (input *ApplicationUpdateInput) ToCompassInput() graphql.ApplicationUpdateInput {
	return graphql.ApplicationUpdateInput(*input)
}
