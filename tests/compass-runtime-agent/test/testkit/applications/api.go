package applications

import "github.com/kyma-incubator/compass/components/director/pkg/graphql"

type APIDefinitionInput graphql.APIDefinitionInput
type AuthInput graphql.AuthInput

func NewAPI(name, description, targetURL string) *APIDefinitionInput {
	api := APIDefinitionInput(graphql.APIDefinitionInput{
		Name:        name,
		Description: &description,
		TargetURL:   targetURL,
	})
	return &api
}

func (in *APIDefinitionInput) ToCompassInput() *graphql.APIDefinitionInput {
	api := graphql.APIDefinitionInput(*in)
	return &api
}

func (in *APIDefinitionInput) WithJsonApiSpec(data *graphql.CLOB) *APIDefinitionInput {
	in.Spec = &graphql.APISpecInput{
		Data:   data,
		Type:   graphql.APISpecTypeOpenAPI,
		Format: graphql.SpecFormatJSON,
	}
	return in
}

func (in *APIDefinitionInput) WithYamlApiSpec(data *graphql.CLOB) *APIDefinitionInput {
	in.Spec = &graphql.APISpecInput{
		Data:   data,
		Type:   graphql.APISpecTypeOpenAPI,
		Format: graphql.SpecFormatYaml,
	}
	return in
}

func (in *APIDefinitionInput) WithXMLApiSpec(data *graphql.CLOB) *APIDefinitionInput {
	in.Spec = &graphql.APISpecInput{
		Data:   data,
		Type:   graphql.APISpecTypeOdata,
		Format: graphql.SpecFormatXML,
	}
	return in
}
