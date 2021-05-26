package director

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/stretchr/testify/assert"

	kymamodel "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/model"
)

const (
	baseAPIId   = "apiId"
	baseAPIName = "awesome api name"
	baseAPIDesc = "so awesome this api description"
	baseAPIURL  = "https://api.url.com"

	baseDocTitle       = "my-docu"
	baseDocDisplayName = "my-docu-display"
	baseDocKind        = "kind-of-cool"

	basePackageId          = "packageId"
	basePackageName        = "packageName"
	basePackageDesc        = "packageDesc"
	basePackageInputSchema = "input schema"
)

// TODO: add test cases for request parameters when they will be supported

func TestApplication_ToApplication(t *testing.T) {

	appId := "abcd"
	appName := "my awesome app"
	providerName := "provider"
	appDesc := "app is so awesome"
	appLabels := map[string]interface{}{
		"appSlice": []string{appName, "app"},
		"app":      "test",
	}

	for _, testCase := range []struct {
		description string
		compassApp  Application
		expectedApp kymamodel.Application
	}{
		{
			description: "not fail if Application is empty",
			expectedApp: kymamodel.Application{
				SystemAuthsIDs: make([]string, 0),
			},
		},
		{
			description: "convert Compass App with auths to internal model",
			compassApp: Application{
				ID:           appId,
				Name:         appName,
				ProviderName: &providerName,
				Description:  &appDesc,
				Labels:       Labels(appLabels),
				Auths: []*graphql.SystemAuth{
					{ID: "1", Auth: &graphql.Auth{Credential: graphql.BasicCredentialData{Password: "password", Username: "user"}}},
					{ID: "2", Auth: &graphql.Auth{Credential: graphql.OAuthCredentialData{ClientSecret: "secret", ClientID: "id"}}},
				},
			},
			expectedApp: kymamodel.Application{
				ID:                  appId,
				Name:                appName,
				ProviderDisplayName: providerName,
				Description:         appDesc,
				Labels:              appLabels,
				SystemAuthsIDs:      []string{"1", "2"},
			},
		},
		{
			description: "convert Compass App using API Packages to internal model",
			compassApp: Application{
				ID:           appId,
				Name:         appName,
				ProviderName: &providerName,
				Description:  &appDesc,
				Labels:       Labels(appLabels),
				Packages: &graphql.PackagePageExt{
					Data: []*graphql.PackageExt{
						fixCompassPackageExt("1"),
						fixCompassPackageExt("2"),
						fixCompassPackageExt("3"),
					},
				},
			},
			expectedApp: kymamodel.Application{
				ID:                  appId,
				Name:                appName,
				ProviderDisplayName: providerName,
				Description:         appDesc,
				Labels:              appLabels,
				SystemAuthsIDs:      make([]string, 0),
				APIPackages: []kymamodel.APIPackage{
					fixInternalAPIPackage("1"),
					fixInternalAPIPackage("2"),
					fixInternalAPIPackage("3"),
				},
			},
		},
		{
			description: "convert Compass App using API Packages to internal model with default instance auth",
			compassApp: Application{
				ID:           appId,
				Name:         appName,
				ProviderName: &providerName,
				Description:  &appDesc,
				Labels:       Labels(appLabels),
				Packages: &graphql.PackagePageExt{
					Data: []*graphql.PackageExt{
						fixCompassPackageWithDefaultInstanceAuth("1", nil),
						fixCompassPackageWithDefaultInstanceAuth("2", &graphql.Auth{
							Credential: &graphql.BasicCredentialData{
								Username: "my-user",
								Password: "my-password",
							},
							AdditionalHeaders:     &graphql.HttpHeaders{"h1": {"v1"}},
							AdditionalQueryParams: &graphql.QueryParams{"q1": {"p1"}},
						}),
						fixCompassPackageWithDefaultInstanceAuth("3", &graphql.Auth{
							Credential: &graphql.OAuthCredentialData{
								ClientID:     "my-client-id",
								ClientSecret: "my-client-secret",
								URL:          "https://test-oauth.com",
							},
							AdditionalHeaders:     &graphql.HttpHeaders{"h1": {"v1"}},
							AdditionalQueryParams: &graphql.QueryParams{"q1": {"p1"}},
						}),
						fixCompassPackageWithDefaultInstanceAuth("4", &graphql.Auth{
							Credential: &graphql.BasicCredentialData{
								Username: "my-user2",
								Password: "my-password2",
							},
							RequestAuth: &graphql.CredentialRequestAuth{
								Csrf: &graphql.CSRFTokenCredentialRequestAuth{
									TokenEndpointURL: "https://csrf.basic.example.com",
								},
							},
						}),
						fixCompassPackageWithDefaultInstanceAuth("5", &graphql.Auth{
							Credential: &graphql.OAuthCredentialData{
								ClientID:     "my-client-id2",
								ClientSecret: "my-client-secret2",
								URL:          "https://test2-oauth.com",
							},
							RequestAuth: &graphql.CredentialRequestAuth{
								Csrf: &graphql.CSRFTokenCredentialRequestAuth{
									TokenEndpointURL: "https://csrf.oauth.example.com",
								},
							},
						}),
					},
				},
			},
			expectedApp: kymamodel.Application{
				ID:                  appId,
				Name:                appName,
				ProviderDisplayName: providerName,
				Description:         appDesc,
				Labels:              appLabels,
				SystemAuthsIDs:      make([]string, 0),
				APIPackages: []kymamodel.APIPackage{
					fixInternalAPIPackageWithInstanceAuth("1", nil),
					fixInternalAPIPackageWithInstanceAuth("2", &kymamodel.Auth{
						Credentials: &kymamodel.Credentials{
							Basic: &kymamodel.Basic{
								Username: "my-user",
								Password: "my-password",
							},
						},
						RequestParameters: &kymamodel.RequestParameters{
							Headers: &map[string][]string{
								"h1": {"v1"},
							},
							QueryParameters: &map[string][]string{
								"q1": {"p1"},
							},
						},
					}),
					fixInternalAPIPackageWithInstanceAuth("3", &kymamodel.Auth{
						Credentials: &kymamodel.Credentials{
							Oauth: &kymamodel.Oauth{
								ClientID:     "my-client-id",
								ClientSecret: "my-client-secret",
								URL:          "https://test-oauth.com",
							},
						},
						RequestParameters: &kymamodel.RequestParameters{
							Headers: &map[string][]string{
								"h1": {"v1"},
							},
							QueryParameters: &map[string][]string{
								"q1": {"p1"},
							},
						},
					}),
					fixInternalAPIPackageWithInstanceAuth("4", &kymamodel.Auth{
						Credentials: &kymamodel.Credentials{
							Basic: &kymamodel.Basic{
								Username: "my-user2",
								Password: "my-password2",
							},
							CSRFInfo: &kymamodel.CSRFInfo{
								TokenEndpointURL: "https://csrf.basic.example.com",
							},
						},
					}),
					fixInternalAPIPackageWithInstanceAuth("5", &kymamodel.Auth{
						Credentials: &kymamodel.Credentials{
							Oauth: &kymamodel.Oauth{
								ClientID:     "my-client-id2",
								ClientSecret: "my-client-secret2",
								URL:          "https://test2-oauth.com",
							},
							CSRFInfo: &kymamodel.CSRFInfo{
								TokenEndpointURL: "https://csrf.oauth.example.com",
							},
						},
					}),
				},
			},
		},
		{
			description: "convert Compass App using API Packages with instance auths to internal model",
			compassApp: Application{
				ID:           appId,
				Name:         appName,
				ProviderName: &providerName,
				Description:  &appDesc,
				Labels:       Labels(appLabels),
				Packages: &graphql.PackagePageExt{
					Data: []*graphql.PackageExt{
						fixCompassPackageExt("1"),
						fixCompassPackageExt("2"),
						fixCompassPackageExt("3"),
					},
				},
			},
			expectedApp: kymamodel.Application{
				ID:                  appId,
				Name:                appName,
				ProviderDisplayName: providerName,
				Description:         appDesc,
				Labels:              appLabels,
				SystemAuthsIDs:      make([]string, 0),
				APIPackages: []kymamodel.APIPackage{
					fixInternalAPIPackage("1"),
					fixInternalAPIPackage("2"),
					fixInternalAPIPackage("3"),
				},
			},
		},
		{
			description: "convert Compass App with empty Package pages",
			compassApp: Application{
				ID:           appId,
				Name:         appName,
				Description:  &appDesc,
				ProviderName: &providerName,
				Labels:       Labels(appLabels),
			},
			expectedApp: kymamodel.Application{
				ID:                  appId,
				Name:                appName,
				Description:         appDesc,
				ProviderDisplayName: providerName,
				Labels:              appLabels,
				SystemAuthsIDs:      make([]string, 0),
			},
		},
		{
			description: "convert Compass App with packages using empty specs",
			compassApp: Application{
				ID:           appId,
				Name:         appName,
				Description:  &appDesc,
				ProviderName: &providerName,
				Packages: &graphql.PackagePageExt{
					Data: []*graphql.PackageExt{
						fixCompassPackageExtWithEmptySpecs("1"),
						fixCompassPackageExtWithEmptySpecs("2"),
						fixCompassPackageExtWithEmptySpecs("3"),
					},
				},
			},
			expectedApp: kymamodel.Application{
				ID:                  appId,
				Name:                appName,
				Description:         appDesc,
				ProviderDisplayName: providerName,
				APIPackages: []kymamodel.APIPackage{
					fixInternalAPIPackageEmptySpecs("1"),
					fixInternalAPIPackageEmptySpecs("2"),
					fixInternalAPIPackageEmptySpecs("3"),
				},
				SystemAuthsIDs: make([]string, 0),
			},
		},
	} {
		t.Run(testCase.description, func(t *testing.T) {
			// when
			internalApp := testCase.compassApp.ToApplication()

			// then
			assert.Equal(t, testCase.expectedApp, internalApp)
		})
	}

}

func fixInternalAPIPackage(suffix string) kymamodel.APIPackage {

	return kymamodel.APIPackage{
		ID:                             basePackageId + suffix,
		Name:                           basePackageName + suffix,
		Description:                    stringPtr(basePackageDesc + suffix),
		InstanceAuthRequestInputSchema: stringPtr(basePackageInputSchema + suffix),
		APIDefinitions: []kymamodel.APIDefinition{
			fixInternalAPIDefinition("1", nil, fixInternalOpenAPISpec()),
			fixInternalAPIDefinition("2", nil, fixInternalOpenAPISpec()),
			fixInternalAPIDefinition("3", nil, fixInternalODataSpec()),
			fixInternalAPIDefinition("4", nil, fixInternalODataSpec()),
		},
		EventDefinitions: []kymamodel.EventAPIDefinition{
			fixInternalEventAPIDefinition("1", fixInternalAsyncAPISpec()),
			fixInternalEventAPIDefinition("2", fixInternalAsyncAPISpec()),
		},
		Documents: []kymamodel.Document{
			fixInternalDocument("1", fixInternalDocumentContent()),
			fixInternalDocument("2", fixInternalDocumentContent()),
		},
	}
}

func fixInternalAPIPackageEmptySpecs(suffix string) kymamodel.APIPackage {
	return kymamodel.APIPackage{
		ID:                             basePackageId + suffix,
		Name:                           basePackageName + suffix,
		Description:                    stringPtr(basePackageDesc + suffix),
		InstanceAuthRequestInputSchema: stringPtr(basePackageInputSchema + suffix),
		APIDefinitions: []kymamodel.APIDefinition{
			fixInternalAPIDefinition("1", nil, nil),
			fixInternalAPIDefinition("2", nil, nil),
			fixInternalAPIDefinition("3", nil, nil),
			fixInternalAPIDefinition("4", nil, nil),
		},
		EventDefinitions: []kymamodel.EventAPIDefinition{
			fixInternalEventAPIDefinition("1", nil),
			fixInternalEventAPIDefinition("2", nil),
		},
		Documents: []kymamodel.Document{
			fixInternalDocument("1", nil),
			fixInternalDocument("2", nil),
			fixInternalDocument("3", nil),
		},
	}
}

func fixInternalAPIDefinition(suffix string, credentials *kymamodel.Credentials, spec *kymamodel.APISpec) kymamodel.APIDefinition {
	return kymamodel.APIDefinition{
		ID:          baseAPIId + suffix,
		Name:        baseAPIName + suffix,
		Description: baseAPIDesc + suffix,
		TargetUrl:   baseAPIURL + suffix,
		APISpec:     spec,
		Credentials: credentials,
	}
}

func fixInternalEventAPIDefinition(suffix string, spec *kymamodel.EventAPISpec) kymamodel.EventAPIDefinition {
	return kymamodel.EventAPIDefinition{
		ID:           baseAPIId + suffix,
		Name:         baseAPIName + suffix,
		Description:  baseAPIDesc + suffix,
		EventAPISpec: spec,
	}
}

func fixInternalDocument(suffix string, data []byte) kymamodel.Document {
	kind := baseDocKind + suffix

	return kymamodel.Document{
		ID:          baseAPIId + suffix,
		Description: baseAPIDesc + suffix,
		Title:       baseDocTitle + suffix,
		DisplayName: baseDocDisplayName + suffix,
		Format:      kymamodel.DocumentFormatMarkdown,
		Kind:        &kind,
		Data:        data,
	}
}

func fixInternalODataSpec() *kymamodel.APISpec {
	return &kymamodel.APISpec{
		Data:   []byte(`OData spec`),
		Type:   kymamodel.APISpecTypeOdata,
		Format: kymamodel.SpecFormatXML,
	}
}

func fixInternalOpenAPISpec() *kymamodel.APISpec {
	return &kymamodel.APISpec{
		Data:   []byte(`Open API spec`),
		Type:   kymamodel.APISpecTypeOpenAPI,
		Format: kymamodel.SpecFormatJSON,
	}
}

func fixInternalAsyncAPISpec() *kymamodel.EventAPISpec {
	return &kymamodel.EventAPISpec{
		Data:   []byte(`Async API spec`),
		Type:   kymamodel.EventAPISpecTypeAsyncAPI,
		Format: kymamodel.SpecFormatYAML,
	}
}

func fixInternalDocumentContent() []byte {
	return []byte(`# Md content`)
}

func fixCompassPackageWithDefaultInstanceAuth(suffix string, defaultInstanceAuth *graphql.Auth) *graphql.PackageExt {
	return &graphql.PackageExt{
		Package:          fixCompassPackage(suffix, defaultInstanceAuth),
		APIDefinitions:   fixAPIDefinitionPageExt(),
		EventDefinitions: fixEventAPIDefinitionPageExt(),
		Documents:        fixDocumentPageExt(),
	}
}

func fixInternalAPIPackageWithInstanceAuth(suffix string, defaultInstanceAuth *kymamodel.Auth) kymamodel.APIPackage {
	return kymamodel.APIPackage{
		ID:                             basePackageId + suffix,
		Name:                           basePackageName + suffix,
		Description:                    stringPtr(basePackageDesc + suffix),
		InstanceAuthRequestInputSchema: stringPtr(basePackageInputSchema + suffix),
		APIDefinitions: []kymamodel.APIDefinition{
			fixInternalAPIDefinition("1", nil, fixInternalOpenAPISpec()),
			fixInternalAPIDefinition("2", nil, fixInternalOpenAPISpec()),
			fixInternalAPIDefinition("3", nil, fixInternalODataSpec()),
			fixInternalAPIDefinition("4", nil, fixInternalODataSpec()),
		},
		EventDefinitions: []kymamodel.EventAPIDefinition{
			fixInternalEventAPIDefinition("1", fixInternalAsyncAPISpec()),
			fixInternalEventAPIDefinition("2", fixInternalAsyncAPISpec()),
		},
		Documents: []kymamodel.Document{
			fixInternalDocument("1", fixInternalDocumentContent()),
			fixInternalDocument("2", fixInternalDocumentContent()),
		},
		DefaultInstanceAuth: defaultInstanceAuth,
	}
}

func fixCompassPackageExt(suffix string) *graphql.PackageExt {

	return &graphql.PackageExt{
		Package:          fixCompassPackage(suffix, nil),
		APIDefinitions:   fixAPIDefinitionPageExt(),
		EventDefinitions: fixEventAPIDefinitionPageExt(),
		Documents:        fixDocumentPageExt(),
	}
}

func fixCompassPackageExtWithEmptySpecs(suffix string) *graphql.PackageExt {
	return &graphql.PackageExt{
		Package:          fixCompassPackage(suffix, nil),
		APIDefinitions:   fixAPIDefinitionPageExtWithEmptyApiSpecs(),
		EventDefinitions: fixEventAPIDefinitionPageExtWithEmptySpecs(),
		Documents:        fixDocumentPageExtWithEmptyDocs(),
	}
}

func fixCompassPackage(suffix string, defaultInstanceAuth *graphql.Auth) graphql.Package {

	return graphql.Package{
		ID:                             basePackageId + suffix,
		Name:                           basePackageName + suffix,
		Description:                    stringPtr(basePackageDesc + suffix),
		InstanceAuthRequestInputSchema: (*graphql.JSONSchema)(stringPtr(basePackageInputSchema + suffix)),
		DefaultInstanceAuth:            defaultInstanceAuth,
	}
}

func fixAPIDefinitionPageExt() graphql.APIDefinitionPageExt {

	return graphql.APIDefinitionPageExt{
		Data: []*graphql.APIDefinitionExt{
			fixCompassAPIDefinitionExt("1", fixCompassOpenAPISpecExt()),
			fixCompassAPIDefinitionExt("2", fixCompassOpenAPISpecExt()),
			fixCompassAPIDefinitionExt("3", fixCompassODataSpecExt()),
			fixCompassAPIDefinitionExt("4", fixCompassODataSpecExt()),
		},
	}
}

func fixAPIDefinitionPageExtWithEmptyApiSpecs() graphql.APIDefinitionPageExt {

	return graphql.APIDefinitionPageExt{
		Data: []*graphql.APIDefinitionExt{
			fixCompassAPIDefinitionExt("1", nil),
			fixCompassAPIDefinitionExt("2", nil),
			fixCompassAPIDefinitionExt("3", nil),
			fixCompassAPIDefinitionExt("4", nil),
		},
	}
}

func fixCompassAPIDefinitionExt(suffix string, spec *graphql.APISpecExt) *graphql.APIDefinitionExt {

	var apiSpec *graphql.APISpec = nil

	if spec != nil {
		apiSpec = &spec.APISpec
	}

	apiDefinition := fixCompassAPIDefinition(suffix, apiSpec)

	return &graphql.APIDefinitionExt{
		APIDefinition: *apiDefinition,
		Spec:          spec,
	}
}

func fixCompassOpenAPISpecExt() *graphql.APISpecExt {

	apiSpec := fixCompassOpenAPISpec()
	return &graphql.APISpecExt{
		APISpec: *apiSpec,
	}
}

func fixCompassODataSpecExt() *graphql.APISpecExt {

	apiSpec := fixCompassODataSpec()
	return &graphql.APISpecExt{
		APISpec: *apiSpec,
	}
}

func fixEventAPIDefinitionPageExt() graphql.EventAPIDefinitionPageExt {

	return graphql.EventAPIDefinitionPageExt{
		Data: []*graphql.EventAPIDefinitionExt{
			fixEventAPIDefinitionExt("1", fixCompassEventAPISpecExt()),
			fixEventAPIDefinitionExt("2", fixCompassEventAPISpecExt()),
		},
	}
}

func fixEventAPIDefinitionPageExtWithEmptySpecs() graphql.EventAPIDefinitionPageExt {

	return graphql.EventAPIDefinitionPageExt{
		Data: []*graphql.EventAPIDefinitionExt{
			fixEventAPIDefinitionExt("1", nil),
			fixEventAPIDefinitionExt("2", nil),
		},
	}
}

func fixCompassEventAPISpecExt() *graphql.EventAPISpecExt {

	eventSpec := fixCompassAsyncAPISpec()

	return &graphql.EventAPISpecExt{
		EventSpec: *eventSpec,
	}
}

func fixEventAPIDefinitionExt(suffix string, extEventSpec *graphql.EventAPISpecExt) *graphql.EventAPIDefinitionExt {

	var eventSpec *graphql.EventSpec = nil

	if extEventSpec != nil {
		eventSpec = &extEventSpec.EventSpec
	}

	eventDefinition := fixCompassEventAPIDefinition(suffix, eventSpec)

	return &graphql.EventAPIDefinitionExt{
		EventDefinition: *eventDefinition,
		Spec:            extEventSpec,
	}
}

func fixDocumentPageExt() graphql.DocumentPageExt {
	return graphql.DocumentPageExt{

		Data: []*graphql.DocumentExt{
			fixCompassDocumentExt("1", fixCompassDocContent()),
			fixCompassDocumentExt("2", fixCompassDocContent()),
		},
	}
}

func fixDocumentPageExtWithEmptyDocs() graphql.DocumentPageExt {
	return graphql.DocumentPageExt{

		Data: []*graphql.DocumentExt{
			fixCompassDocumentExt("1", nil),
			fixCompassDocumentExt("2", nil),
			fixCompassDocumentExt("3", nil),
		},
	}
}

func fixCompassDocumentExt(suffix string, content *graphql.CLOB) *graphql.DocumentExt {

	document := fixCompassDocument(suffix, content)

	return &graphql.DocumentExt{
		Document: *document,
	}
}

func fixCompassAPIDefinition(suffix string, spec *graphql.APISpec) *graphql.APIDefinition {
	desc := baseAPIDesc + suffix

	return &graphql.APIDefinition{
		ID:          baseAPIId + suffix,
		Name:        baseAPIName + suffix,
		Description: &desc,
		TargetURL:   baseAPIURL + suffix,
		Spec:        spec,
	}
}

func fixCompassEventAPIDefinition(suffix string, spec *graphql.EventSpec) *graphql.EventDefinition {
	desc := baseAPIDesc + suffix

	return &graphql.EventDefinition{
		ID:          baseAPIId + suffix,
		Name:        baseAPIName + suffix,
		Description: &desc,
		Spec:        spec,
	}
}

func fixCompassDocument(suffix string, data *graphql.CLOB) *graphql.Document {
	kind := baseDocKind + suffix

	return &graphql.Document{
		ID:          baseAPIId + suffix,
		Description: baseAPIDesc + suffix,
		Title:       baseDocTitle + suffix,
		DisplayName: baseDocDisplayName + suffix,
		Format:      graphql.DocumentFormatMarkdown,
		Kind:        &kind,
		Data:        data,
	}
}

func fixCompassODataSpec() *graphql.APISpec {
	data := graphql.CLOB(`OData spec`)

	return &graphql.APISpec{
		Data:   &data,
		Type:   graphql.APISpecTypeOdata,
		Format: graphql.SpecFormatXML,
	}
}

func fixCompassOpenAPISpec() *graphql.APISpec {
	data := graphql.CLOB(`Open API spec`)

	return &graphql.APISpec{
		Data:   &data,
		Type:   graphql.APISpecTypeOpenAPI,
		Format: graphql.SpecFormatJSON,
	}
}

func fixCompassAsyncAPISpec() *graphql.EventSpec {
	data := graphql.CLOB(`Async API spec`)

	return &graphql.EventSpec{
		Data:   &data,
		Type:   graphql.EventSpecTypeAsyncAPI,
		Format: graphql.SpecFormatYaml,
	}
}

func fixCompassDocContent() *graphql.CLOB {
	data := graphql.CLOB(`# Md content`)

	return &data
}

func stringPtr(str string) *string {
	return &str
}
