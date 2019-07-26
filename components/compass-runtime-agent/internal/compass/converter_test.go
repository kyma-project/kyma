package compass

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/synchronization"
)

const (
	clientId     = "oauth-client-id"
	clientSecret = "oauth-client-secret"
	oauthURL     = "https://give.me.token/token"

	username = "basic-username"
	password = "basic-password"

	csrfTokenURL = "http://csrf.url.com/token"

	baseAPIId   = "apiId"
	baseAPIName = "awesome api name"
	baseAPIDesc = "so awesome this api description"
	baseAPIURL  = "https://api.url.com"

	baseDocTitle       = "my-docu"
	baseDocDisplayName = "my-docu-display"
	baseDocKind        = "kind-of-cool"
)

// TODO: add test cases for request parameters when they will be supported

func TestApplication_ToApplication(t *testing.T) {

	appId := "abcd"
	appName := "my awesome app"
	appDesc := "app is so awesome"
	appLabels := map[string][]string{
		"app": {appName, "app"},
	}

	for _, testCase := range []struct {
		description string
		compassApp  Application
		expectedApp synchronization.Application
	}{
		{
			description: "not fail if Application is empty",
		},
		{
			description: "convert Compass App to internal model",
			compassApp: Application{
				ID:          appId,
				Name:        appName,
				Description: &appDesc,
				Labels:      Labels(appLabels),
				APIs: &graphql.APIDefinitionPage{
					Data: []*graphql.APIDefinition{
						fixCompassAPIDefinition("1", fixCompassOauthAuth(nil), fixCompassOpenAPISpec()),
						fixCompassAPIDefinition("2", fixCompassBasicAuthAuth(nil), fixCompassODataSpec()),
						fixCompassAPIDefinition("3", fixCompassBasicAuthAuth(fixCompassRequestAuth()), fixCompassODataSpec()),
						fixCompassAPIDefinition("4", nil, fixCompassODataSpec()),
						fixCompassAPIDefinition("5", nil, nil),
					},
				},
				EventAPIs: &graphql.EventAPIDefinitionPage{
					Data: []*graphql.EventAPIDefinition{
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
			},
			expectedApp: synchronization.Application{
				ID:          appId,
				Name:        appName,
				Description: &appDesc,
				Labels:      appLabels,
				APIs: []synchronization.APIDefinition{
					fixInternalAPIDefinition("1", fixInternalOauthCredentials(nil), fixInternalOpenAPISpec()),
					fixInternalAPIDefinition("2", fixInternalBasicAuthCredentials(nil), fixInternalODataSpec()),
					fixInternalAPIDefinition("3", fixInternalBasicAuthCredentials(fixInternalCSRFInfo()), fixInternalODataSpec()),
					fixInternalAPIDefinition("4", nil, fixInternalODataSpec()),
					fixInternalAPIDefinition("5", nil, nil),
				},
				EventAPIs: []synchronization.EventAPIDefinition{
					fixInternalEventAPIDefinition("1", fixInternalAsyncAPISpec()),
					fixInternalEventAPIDefinition("2", fixInternalAsyncAPISpec()),
					fixInternalEventAPIDefinition("3", nil),
				},
				Documents: []synchronization.Document{
					fixInternalDocument("1", fixInternalDocumentContent()),
					fixInternalDocument("2", fixInternalDocumentContent()),
					fixInternalDocument("3", nil),
				},
			},
		},
		{
			description: "convert Compass App with empty pages",
			compassApp: Application{
				ID:          appId,
				Name:        appName,
				Description: &appDesc,
				APIs: &graphql.APIDefinitionPage{
					Data: []*graphql.APIDefinition{
						{},
					},
				},
				EventAPIs: &graphql.EventAPIDefinitionPage{
					Data: []*graphql.EventAPIDefinition{
						{},
					},
				},
				Documents: &graphql.DocumentPage{
					Data: []*graphql.Document{
						{},
					},
				},
			},
			expectedApp: synchronization.Application{
				ID:          appId,
				Name:        appName,
				Description: &appDesc,
				APIs: []synchronization.APIDefinition{
					{},
				},
				EventAPIs: []synchronization.EventAPIDefinition{
					{},
				},
				Documents: []synchronization.Document{
					{},
				},
			},
		},
		{
			description: "convert Compass App with empty apis",
			compassApp: Application{
				ID:          appId,
				Name:        appName,
				Description: &appDesc,
				APIs: &graphql.APIDefinitionPage{
					Data: nil,
				},
				EventAPIs: &graphql.EventAPIDefinitionPage{
					Data: nil,
				},
				Documents: &graphql.DocumentPage{
					Data: nil,
				},
			},
			expectedApp: synchronization.Application{
				ID:          appId,
				Name:        appName,
				Description: &appDesc,
				APIs:        []synchronization.APIDefinition{},
				EventAPIs:   []synchronization.EventAPIDefinition{},
				Documents:   []synchronization.Document{},
			},
		},
		{
			description: "set empty credentials when unsupported credentials input",
			compassApp: Application{
				ID:          appId,
				Name:        appName,
				Description: &appDesc,
				Labels:      Labels(appLabels),
				APIs: &graphql.APIDefinitionPage{
					Data: []*graphql.APIDefinition{
						fixCompassAPIDefinition("1", fixCompassUnsupportedCredentialsAuth(), fixCompassOpenAPISpec()),
					},
				},
			},
			expectedApp: synchronization.Application{
				ID:          appId,
				Name:        appName,
				Description: &appDesc,
				Labels:      appLabels,
				APIs: []synchronization.APIDefinition{
					fixInternalAPIDefinition("1", nil, fixInternalOpenAPISpec()),
				},
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

func fixCompassUnsupportedCredentialsAuth() *graphql.RuntimeAuth {
	return &graphql.RuntimeAuth{
		RuntimeID: runtimeId,
		Auth: &graphql.Auth{
			Credential: UnsupportedCredentials{},
		},
	}
}

func fixInternalAPIDefinition(suffix string, credentials *synchronization.Credentials, spec *synchronization.APISpec) synchronization.APIDefinition {
	return synchronization.APIDefinition{
		ID:          baseAPIId + suffix,
		Name:        baseAPIName + suffix,
		Description: baseAPIDesc + suffix,
		TargetUrl:   baseAPIURL + suffix,
		Credentials: credentials,
		APISpec:     spec,
	}
}

func fixInternalEventAPIDefinition(suffix string, spec *synchronization.EventAPISpec) synchronization.EventAPIDefinition {
	return synchronization.EventAPIDefinition{
		ID:           baseAPIId + suffix,
		Name:         baseAPIName + suffix,
		Description:  baseAPIDesc + suffix,
		EventAPISpec: spec,
	}
}

func fixInternalDocument(suffix string, data []byte) synchronization.Document {
	kind := baseDocKind + suffix

	return synchronization.Document{
		ID:          baseAPIId + suffix,
		Description: baseAPIDesc + suffix,
		Title:       baseDocTitle + suffix,
		DisplayName: baseDocDisplayName + suffix,
		Format:      synchronization.DocumentFormatMarkdown,
		Kind:        &kind,
		Data:        data,
	}
}

func fixInternalOauthCredentials(csrf *synchronization.CSRFInfo) *synchronization.Credentials {
	return &synchronization.Credentials{
		Oauth: &synchronization.Oauth{
			URL:          oauthURL,
			ClientID:     clientId,
			ClientSecret: clientSecret,
		},
		CSRFInfo: csrf,
	}
}

func fixInternalBasicAuthCredentials(csrf *synchronization.CSRFInfo) *synchronization.Credentials {
	return &synchronization.Credentials{
		Basic: &synchronization.Basic{
			Username: username,
			Password: password,
		},
		CSRFInfo: csrf,
	}
}

func fixInternalCSRFInfo() *synchronization.CSRFInfo {
	return &synchronization.CSRFInfo{
		TokenEndpointURL: csrfTokenURL,
	}
}

func fixInternalODataSpec() *synchronization.APISpec {
	return &synchronization.APISpec{
		Data: []byte(`OData spec`),
		Type: synchronization.APISpecTypeOdata,
	}
}

func fixInternalOpenAPISpec() *synchronization.APISpec {
	return &synchronization.APISpec{
		Data: []byte(`Open API spec`),
		Type: synchronization.APISpecTypeOpenAPI,
	}
}

func fixInternalAsyncAPISpec() *synchronization.EventAPISpec {
	return &synchronization.EventAPISpec{
		Data: []byte(`Async API spec`),
		Type: synchronization.EventAPISpecTypeAsyncAPI,
	}
}

func fixInternalDocumentContent() []byte {
	return []byte(`# Md content`)
}

func fixCompassAPIDefinition(suffix string, auth *graphql.RuntimeAuth, spec *graphql.APISpec) *graphql.APIDefinition {
	desc := baseAPIDesc + suffix

	return &graphql.APIDefinition{
		ID:          baseAPIId + suffix,
		Name:        baseAPIName + suffix,
		Description: &desc,
		TargetURL:   baseAPIURL + suffix,
		Auth:        auth,
		Spec:        spec,
	}
}

func fixCompassEventAPIDefinition(suffix string, spec *graphql.EventAPISpec) *graphql.EventAPIDefinition {
	desc := baseAPIDesc + suffix

	return &graphql.EventAPIDefinition{
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

func fixCompassOauthAuth(requestAuth *graphql.CredentialRequestAuth) *graphql.RuntimeAuth {
	return &graphql.RuntimeAuth{
		RuntimeID: runtimeId,
		Auth: &graphql.Auth{
			Credential: graphql.OAuthCredentialData{
				URL:          oauthURL,
				ClientID:     clientId,
				ClientSecret: clientSecret,
			},
			RequestAuth: requestAuth,
		},
	}
}

func fixCompassBasicAuthAuth(requestAuth *graphql.CredentialRequestAuth) *graphql.RuntimeAuth {
	return &graphql.RuntimeAuth{
		RuntimeID: runtimeId,
		Auth: &graphql.Auth{
			Credential: graphql.BasicCredentialData{
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
		Data: &data,
		Type: graphql.APISpecTypeOdata,
	}
}

func fixCompassOpenAPISpec() *graphql.APISpec {
	data := graphql.CLOB(`Open API spec`)

	return &graphql.APISpec{
		Data: &data,
		Type: graphql.APISpecTypeOpenAPI,
	}
}

func fixCompassAsyncAPISpec() *graphql.EventAPISpec {
	data := graphql.CLOB(`Async API spec`)

	return &graphql.EventAPISpec{
		Data: &data,
		Type: graphql.EventAPISpecTypeAsyncAPI,
	}
}

func fixCompassDocContent() *graphql.CLOB {
	data := graphql.CLOB(`# Md content`)

	return &data
}
