package director

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/stretchr/testify/assert"

	kymamodel "kyma-project.io/compass-runtime-agent/internal/kyma/model"
)

const (
	clientId     = "oauth-client-id"
	clientSecret = "oauth-client-secret"
	oauthURL     = "https://give.me.token/token"

	username = "basic-username"
	password = "basic-password"

	baseAPIId   = "apiId"
	baseAPIName = "awesome api name"
	baseAPIDesc = "so awesome this api description"
	baseAPIURL  = "https://api.url.com"

	baseDocTitle       = "my-docu"
	baseDocDisplayName = "my-docu-display"
	baseDocKind        = "kind-of-cool"

	csrfTokenURL = "http://csrf.url.com/token"

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
			description: "convert Compass App to internal model",
			compassApp: Application{
				ID:           appId,
				Name:         appName,
				ProviderName: &providerName,
				Description:  &appDesc,
				Labels:       Labels(appLabels),
				APIDefinitions: &graphql.APIDefinitionPage{
					Data: []*graphql.APIDefinition{
						fixCompassAPIDefinition("1", fixCompassOauthAuth(nil), fixCompassOpenAPISpec()),
						fixCompassAPIDefinition("2", fixCompassBasicAuthAuth(nil), fixCompassODataSpec()),
						fixCompassAPIDefinition("3", fixCompassBasicAuthAuth(fixCompassRequestAuth()), fixCompassODataSpec()),
						fixCompassAPIDefinition("4", nil, fixCompassODataSpec()),
						fixCompassAPIDefinition("5", nil, nil),
					},
				},
				EventDefinitions: &graphql.EventDefinitionPage{
					Data: []*graphql.EventDefinition{
						fixCompassEventAPIDefinition("1", fixCompassAsyncAPISpec()),
						fixCompassEventAPIDefinition("2", fixCompassAsyncAPISpec()),
						fixCompassEventAPIDefinition("3", nil),
					},
				},
				Documents: &graphql.DocumentPage{
					Data: []*graphql.Document{
						fixCompassDocument("1", fixCompassDocContent()),
						fixCompassDocument("2", fixCompassDocContent()),
						fixCompassDocument("3", nil),
					},
				},
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
				APIs: []kymamodel.APIDefinition{
					fixInternalAPIDefinition("1", fixInternalOauthCredentials(nil), fixInternalOpenAPISpec()),
					fixInternalAPIDefinition("2", fixInternalBasicAuthCredentials(nil), fixInternalODataSpec()),
					fixInternalAPIDefinition("3", fixInternalBasicAuthCredentials(fixInternalCSRFInfo()), fixInternalODataSpec()),
					fixInternalAPIDefinition("4", nil, fixInternalODataSpec()),
					fixInternalAPIDefinition("5", nil, nil),
				},
				EventAPIs: []kymamodel.EventAPIDefinition{
					fixInternalEventAPIDefinition("1", fixInternalAsyncAPISpec()),
					fixInternalEventAPIDefinition("2", fixInternalAsyncAPISpec()),
					fixInternalEventAPIDefinition("3", nil),
				},
				Documents: []kymamodel.Document{
					fixInternalDocument("1", fixInternalDocumentContent()),
					fixInternalDocument("2", fixInternalDocumentContent()),
					fixInternalDocument("3", nil),
				},
				SystemAuthsIDs: []string{"1", "2"},
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
			description: "convert Compass App with empty pages",
			compassApp: Application{
				ID:          appId,
				Name:        appName,
				Description: &appDesc,
				APIDefinitions: &graphql.APIDefinitionPage{
					Data: []*graphql.APIDefinition{
						{},
					},
				},
				EventDefinitions: &graphql.EventDefinitionPage{
					Data: []*graphql.EventDefinition{
						{},
					},
				},
				Documents: &graphql.DocumentPage{
					Data: []*graphql.Document{
						{},
					},
				},
			},
			expectedApp: kymamodel.Application{
				ID:                  appId,
				Name:                appName,
				Description:         appDesc,
				ProviderDisplayName: "",
				APIs: []kymamodel.APIDefinition{
					{},
				},
				EventAPIs: []kymamodel.EventAPIDefinition{
					{},
				},
				Documents: []kymamodel.Document{
					{},
				},
				SystemAuthsIDs: make([]string, 0),
			},
		},
		{
			description: "convert Compass App with empty apis",
			compassApp: Application{
				ID:          appId,
				Name:        appName,
				Description: &appDesc,
				APIDefinitions: &graphql.APIDefinitionPage{
					Data: nil,
				},
				EventDefinitions: &graphql.EventDefinitionPage{
					Data: nil,
				},
				Documents: &graphql.DocumentPage{
					Data: nil,
				},
			},
			expectedApp: kymamodel.Application{
				ID:             appId,
				Name:           appName,
				Description:    appDesc,
				APIs:           []kymamodel.APIDefinition{},
				EventAPIs:      []kymamodel.EventAPIDefinition{},
				Documents:      []kymamodel.Document{},
				SystemAuthsIDs: make([]string, 0),
			},
		},
		{
			description: "convert Compass App with empty specs",
			compassApp: Application{
				ID:          appId,
				Name:        appName,
				Description: &appDesc,
				APIDefinitions: &graphql.APIDefinitionPage{
					Data: []*graphql.APIDefinition{
						fixCompassAPIDefinition("1", fixCompassOauthAuth(nil), &graphql.APISpec{Data: nil}),
					},
				},
				EventDefinitions: &graphql.EventDefinitionPage{
					Data: []*graphql.EventDefinition{
						fixCompassEventAPIDefinition("1", &graphql.EventSpec{Data: nil}),
					},
				},
				Documents: &graphql.DocumentPage{
					Data: []*graphql.Document{
						fixCompassDocument("1", nil),
					},
				},
			},
			expectedApp: kymamodel.Application{
				ID:          appId,
				Name:        appName,
				Description: appDesc,
				APIs: []kymamodel.APIDefinition{
					fixInternalAPIDefinition("1", fixInternalOauthCredentials(nil), &kymamodel.APISpec{}),
				},
				EventAPIs: []kymamodel.EventAPIDefinition{
					fixInternalEventAPIDefinition("1", &kymamodel.EventAPISpec{}),
				},
				Documents: []kymamodel.Document{
					fixInternalDocument("1", nil),
				},
				SystemAuthsIDs: make([]string, 0),
			},
		},
		{
			description: "set empty credentials when unsupported credentials input",
			compassApp: Application{
				ID:          appId,
				Name:        appName,
				Description: &appDesc,
				Labels:      Labels(appLabels),
				APIDefinitions: &graphql.APIDefinitionPage{
					Data: []*graphql.APIDefinition{
						fixCompassAPIDefinition("1", fixCompassUnsupportedCredentialsAuth(), fixCompassOpenAPISpec()),
					},
				},
			},
			expectedApp: kymamodel.Application{
				ID:          appId,
				Name:        appName,
				Description: appDesc,
				Labels:      appLabels,
				APIs: []kymamodel.APIDefinition{
					fixInternalAPIDefinition("1", nil, fixInternalOpenAPISpec()),
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

type UnsupportedCredentials struct{}

func (UnsupportedCredentials) IsCredentialData() {}

func fixCompassUnsupportedCredentialsAuth() *graphql.APIRuntimeAuth {
	return &graphql.APIRuntimeAuth{
		RuntimeID: runtimeId,
		Auth: &graphql.Auth{
			Credential: UnsupportedCredentials{},
		},
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
			fixInternalEventAPIDefinition("3", nil),
		},
		Documents: []kymamodel.Document{
			fixInternalDocument("1", fixInternalDocumentContent()),
			fixInternalDocument("2", fixInternalDocumentContent()),
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

func fixInternalOauthCredentials(csrf *kymamodel.CSRFInfo) *kymamodel.Credentials {
	return &kymamodel.Credentials{
		Oauth: &kymamodel.Oauth{
			URL:          oauthURL,
			ClientID:     clientId,
			ClientSecret: clientSecret,
		},
		CSRFInfo: csrf,
	}
}

func fixInternalBasicAuthCredentials(csrf *kymamodel.CSRFInfo) *kymamodel.Credentials {
	return &kymamodel.Credentials{
		Basic: &kymamodel.Basic{
			Username: username,
			Password: password,
		},
		CSRFInfo: csrf,
	}
}

func fixInternalCSRFInfo() *kymamodel.CSRFInfo {
	return &kymamodel.CSRFInfo{
		TokenEndpointURL: csrfTokenURL,
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

func fixCompassPackageExt(suffix string) *graphql.PackageExt {

	return &graphql.PackageExt{
		Package:          fixCompassPackage(suffix),
		APIDefinitions:   fixAPIDefinitionPageExt(),
		EventDefinitions: fixEventAPIDefinitionPageExt(),
		Documents:        fixDocumentPageExt(),
	}
}

func fixCompassPackage(suffix string) graphql.Package {

	return graphql.Package{
		ID:                             basePackageId + suffix,
		Name:                           basePackageName + suffix,
		Description:                    stringPtr(basePackageDesc + suffix),
		InstanceAuthRequestInputSchema: (*graphql.JSONSchema)(stringPtr(basePackageInputSchema + suffix)),
		DefaultInstanceAuth:            nil,
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

func fixCompassAPIDefinitionExt(suffix string, spec *graphql.APISpecExt) *graphql.APIDefinitionExt {

	var apiSpec *graphql.APISpec = nil

	if spec != nil {
		apiSpec = &spec.APISpec
	}

	apiDefinition := fixCompassAPIDefinition(suffix, nil, apiSpec)

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
			fixEventAPIDefinitionExt("3", nil),
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

func fixCompassAPIDefinition(suffix string, auth *graphql.APIRuntimeAuth, spec *graphql.APISpec) *graphql.APIDefinition {
	desc := baseAPIDesc + suffix

	var defaultAuth *graphql.Auth
	if auth != nil {
		defaultAuth = auth.Auth
	}

	return &graphql.APIDefinition{
		ID:          baseAPIId + suffix,
		Name:        baseAPIName + suffix,
		Description: &desc,
		TargetURL:   baseAPIURL + suffix,
		DefaultAuth: defaultAuth, // TODO: can be removed after switching to use Auth
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

func fixCompassOauthAuth(requestAuth *graphql.CredentialRequestAuth) *graphql.APIRuntimeAuth {
	return &graphql.APIRuntimeAuth{
		RuntimeID: runtimeId,
		Auth: &graphql.Auth{
			Credential: &graphql.OAuthCredentialData{
				URL:          oauthURL,
				ClientID:     clientId,
				ClientSecret: clientSecret,
			},
			RequestAuth: requestAuth,
		},
	}
}

func fixCompassBasicAuthAuth(requestAuth *graphql.CredentialRequestAuth) *graphql.APIRuntimeAuth {
	return &graphql.APIRuntimeAuth{
		RuntimeID: runtimeId,
		Auth: &graphql.Auth{
			Credential: &graphql.BasicCredentialData{
				Username: username,
				Password: password,
			},
			RequestAuth: requestAuth,
		},
	}
}

func fixCompassRequestAuth() *graphql.CredentialRequestAuth {
	return &graphql.CredentialRequestAuth{
		Csrf: &graphql.CSRFTokenCredentialRequestAuth{
			TokenEndpointURL: csrfTokenURL,
		},
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
