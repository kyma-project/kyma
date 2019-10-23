package applications

import "github.com/kyma-incubator/compass/components/director/pkg/graphql"

type ApplicationCreateInput graphql.ApplicationCreateInput

func NewApplication(name, description string, labels map[string]interface{}) *ApplicationCreateInput {
	appLabels := graphql.Labels(labels)

	app := ApplicationCreateInput(graphql.ApplicationCreateInput{
		Name:        name,
		Description: &description,
		Labels:      &appLabels,
		Apis:        nil,
		EventAPIs:   nil,
		Documents:   nil,
	})

	return &app
}

func (input *ApplicationCreateInput) ToCompassInput() graphql.ApplicationCreateInput {
	return graphql.ApplicationCreateInput(*input)
}

func (input *ApplicationCreateInput) WithAPIs(apis []*APIDefinitionInput) *ApplicationCreateInput {
	compassAPIs := make([]*graphql.APIDefinitionInput, len(apis))
	for i, api := range apis {
		compassAPIs[i] = api.ToCompassInput()
	}

	input.Apis = compassAPIs

	return input
}

func (input *ApplicationCreateInput) WithEventAPIs(apis []*EventAPIDefinitionInput) *ApplicationCreateInput {
	compassAPIs := make([]*graphql.EventAPIDefinitionInput, len(apis))
	for i, api := range apis {
		compassAPIs[i] = api.ToCompassInput()
	}

	input.EventAPIs = compassAPIs

	return input
}

type ApplicationUpdateInput graphql.ApplicationUpdateInput

func NewApplicationUpdateInput(name, description string) *ApplicationUpdateInput {
	appUpdateInput := ApplicationUpdateInput(graphql.ApplicationUpdateInput{
		Name:        name,
		Description: &description,
	})

	return &appUpdateInput
}

func (input *ApplicationUpdateInput) ToCompassInput() graphql.ApplicationUpdateInput {
	return graphql.ApplicationUpdateInput(*input)
}
