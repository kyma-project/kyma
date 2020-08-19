package applications

import "github.com/kyma-incubator/compass/components/director/pkg/graphql"

type ApplicationRegisterInput graphql.ApplicationRegisterInput

func NewApplication(name, providerName, description string, labels map[string]interface{}) *ApplicationRegisterInput {
	appLabels := graphql.Labels(labels)

	app := ApplicationRegisterInput(graphql.ApplicationRegisterInput{
		Name:         name,
		ProviderName: &providerName,
		Description:  &description,
		Labels:       &appLabels,
	})

	return &app
}

func (input *ApplicationRegisterInput) ToCompassInput() graphql.ApplicationRegisterInput {
	return graphql.ApplicationRegisterInput(*input)
}

func (input *ApplicationRegisterInput) WithAPIPackages(packages ...*APIPackageInput) *ApplicationRegisterInput {
	apiPackages := make([]*graphql.PackageCreateInput, len(packages))
	for i, pkg := range packages {
		apiPackages[i] = pkg.ToCompassInput()
	}

	input.Packages = apiPackages

	return input
}

type APIPackageInput graphql.PackageCreateInput

func NewAPIPackage(name, description string) *APIPackageInput {
	apiPackage := APIPackageInput(graphql.PackageCreateInput{
		Name:                name,
		Description:         &description,
		DefaultInstanceAuth: &graphql.AuthInput{},
		APIDefinitions:      nil,
		EventDefinitions:    nil,
		Documents:           nil,
	})

	return &apiPackage
}

func (input *APIPackageInput) ToCompassInput() *graphql.PackageCreateInput {
	pkg := graphql.PackageCreateInput(*input)
	return &pkg
}

func (input *APIPackageInput) WithAPIDefinitions(apis []*APIDefinitionInput) *APIPackageInput {
	compassAPIs := make([]*graphql.APIDefinitionInput, len(apis))
	for i, api := range apis {
		compassAPIs[i] = api.ToCompassInput()
	}

	input.APIDefinitions = compassAPIs

	return input
}

func (input *APIPackageInput) WithEventDefinitions(apis []*EventDefinitionInput) *APIPackageInput {
	compassAPIs := make([]*graphql.EventDefinitionInput, len(apis))
	for i, api := range apis {
		compassAPIs[i] = api.ToCompassInput()
	}

	input.EventDefinitions = compassAPIs

	return input
}

func (in *APIPackageInput) WithAuth(auth *AuthInput) *APIPackageInput {
	in.DefaultInstanceAuth = auth.ToCompassInput()
	return in
}

type APIPackageUpdateInput graphql.PackageUpdateInput

func NewAPIPackageUpdateInput(name, description string, auth *graphql.AuthInput) *APIPackageUpdateInput {
	apiPackage := APIPackageUpdateInput(graphql.PackageUpdateInput{
		Name:                name,
		Description:         &description,
		DefaultInstanceAuth: auth,
	})

	return &apiPackage
}

func (input *APIPackageUpdateInput) ToCompassInput() graphql.PackageUpdateInput {
	return graphql.PackageUpdateInput(*input)
}

type ApplicationUpdateInput graphql.ApplicationUpdateInput

func NewApplicationUpdateInput(providerName, description string) *ApplicationUpdateInput {
	appUpdateInput := ApplicationUpdateInput(graphql.ApplicationUpdateInput{
		ProviderName: &providerName,
		Description:  &description,
	})

	return &appUpdateInput
}

func (input *ApplicationUpdateInput) ToCompassInput() graphql.ApplicationUpdateInput {
	return graphql.ApplicationUpdateInput(*input)
}
