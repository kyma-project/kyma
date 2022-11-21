package director

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"

	kymaModel "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/model"
)

const (
	baseAPIId   = "apiId"
	baseAPIName = "awesome api name"
	baseAPIDesc = "so awesome this api description"
	baseAPIURL  = "https://api.url.com"

	baseDocTitle       = "my-docu"
	baseDocDisplayName = "my-docu-display"
	baseDocKind        = "kind-of-cool"

	baseBundleID          = "bundleID"
	baseBundleName        = "bundleName"
	baseBundleDesc        = "bundleDesc"
	baseBundleInputSchema = "input schema"
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
		expectedApp kymaModel.Application
	}{
		{
			description: "not fail if Application is empty",
			expectedApp: kymaModel.Application{
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
				Labels:       appLabels,
				Auths: []*graphql.AppSystemAuth{
					{ID: "1", Auth: &graphql.Auth{Credential: graphql.BasicCredentialData{Password: "password", Username: "user"}}},
					{ID: "2", Auth: &graphql.Auth{Credential: graphql.OAuthCredentialData{ClientSecret: "secret", ClientID: "id"}}},
				},
			},
			expectedApp: kymaModel.Application{
				ID:                  appId,
				Name:                appName,
				ProviderDisplayName: providerName,
				Description:         appDesc,
				Labels:              appLabels,
				SystemAuthsIDs:      []string{"1", "2"},
			},
		},
		{
			description: "convert Compass App using API Bundles to internal model",
			compassApp: Application{
				ID:           appId,
				Name:         appName,
				ProviderName: &providerName,
				Description:  &appDesc,
				Labels:       appLabels,
				Bundles: &graphql.BundlePageExt{
					Data: []*graphql.BundleExt{
						fixCompassBundleExt("1"),
						fixCompassBundleExt("2"),
						fixCompassBundleExt("3"),
					},
				},
			},
			expectedApp: kymaModel.Application{
				ID:                  appId,
				Name:                appName,
				ProviderDisplayName: providerName,
				Description:         appDesc,
				Labels:              appLabels,
				SystemAuthsIDs:      make([]string, 0),
				ApiBundles: []kymaModel.APIBundle{
					fixInternalAPIBundle("1"),
					fixInternalAPIBundle("2"),
					fixInternalAPIBundle("3"),
				},
			},
		},
		{
			description: "convert Compass App using API Bundles to internal model with default instance auth",
			compassApp: Application{
				ID:           appId,
				Name:         appName,
				ProviderName: &providerName,
				Description:  &appDesc,
				Labels:       appLabels,
				Bundles: &graphql.BundlePageExt{
					Data: []*graphql.BundleExt{
						fixCompassBundleWithDefaultInstanceAuth("1", nil),
						fixCompassBundleWithDefaultInstanceAuth("2", &graphql.Auth{
							Credential: &graphql.BasicCredentialData{
								Username: "my-user",
								Password: "my-password",
							},
							AdditionalHeaders:     graphql.HTTPHeaders{"h1": {"v1"}},
							AdditionalQueryParams: graphql.QueryParams{"q1": {"p1"}},
						}),
						fixCompassBundleWithDefaultInstanceAuth("3", &graphql.Auth{
							Credential: &graphql.OAuthCredentialData{
								ClientID:     "my-client-id",
								ClientSecret: "my-client-secret",
								URL:          "https://test-oauth.com",
							},
							AdditionalHeaders:     graphql.HTTPHeaders{"h1": {"v1"}},
							AdditionalQueryParams: graphql.QueryParams{"q1": {"p1"}},
						}),
						fixCompassBundleWithDefaultInstanceAuth("4", &graphql.Auth{
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
						fixCompassBundleWithDefaultInstanceAuth("5", &graphql.Auth{
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
			expectedApp: kymaModel.Application{
				ID:                  appId,
				Name:                appName,
				ProviderDisplayName: providerName,
				Description:         appDesc,
				Labels:              appLabels,
				SystemAuthsIDs:      make([]string, 0),
				ApiBundles: []kymaModel.APIBundle{
					fixInternalAPIBundleWithInstanceAuth("1", nil),
					fixInternalAPIBundleWithInstanceAuth("2", &kymaModel.Auth{
						Credentials: &kymaModel.Credentials{
							Basic: &kymaModel.Basic{
								Username: "my-user",
								Password: "my-password",
							},
						},
						RequestParameters: &kymaModel.RequestParameters{
							Headers: &map[string][]string{
								"h1": {"v1"},
							},
							QueryParameters: &map[string][]string{
								"q1": {"p1"},
							},
						},
					}),
					fixInternalAPIBundleWithInstanceAuth("3", &kymaModel.Auth{
						Credentials: &kymaModel.Credentials{
							Oauth: &kymaModel.Oauth{
								ClientID:     "my-client-id",
								ClientSecret: "my-client-secret",
								URL:          "https://test-oauth.com",
							},
						},
						RequestParameters: &kymaModel.RequestParameters{
							Headers: &map[string][]string{
								"h1": {"v1"},
							},
							QueryParameters: &map[string][]string{
								"q1": {"p1"},
							},
						},
					}),
					fixInternalAPIBundleWithInstanceAuth("4", &kymaModel.Auth{
						Credentials: &kymaModel.Credentials{
							Basic: &kymaModel.Basic{
								Username: "my-user2",
								Password: "my-password2",
							},
							CSRFInfo: &kymaModel.CSRFInfo{
								TokenEndpointURL: "https://csrf.basic.example.com",
							},
						},
					}),
					fixInternalAPIBundleWithInstanceAuth("5", &kymaModel.Auth{
						Credentials: &kymaModel.Credentials{
							Oauth: &kymaModel.Oauth{
								ClientID:     "my-client-id2",
								ClientSecret: "my-client-secret2",
								URL:          "https://test2-oauth.com",
							},
							CSRFInfo: &kymaModel.CSRFInfo{
								TokenEndpointURL: "https://csrf.oauth.example.com",
							},
						},
					}),
				},
			},
		},
		{
			description: "convert Compass App using API Bundles with instance auths to internal model",
			compassApp: Application{
				ID:           appId,
				Name:         appName,
				ProviderName: &providerName,
				Description:  &appDesc,
				Labels:       appLabels,
				Bundles: &graphql.BundlePageExt{
					Data: []*graphql.BundleExt{
						fixCompassBundleExt("1"),
						fixCompassBundleExt("2"),
						fixCompassBundleExt("3"),
					},
				},
			},
			expectedApp: kymaModel.Application{
				ID:                  appId,
				Name:                appName,
				ProviderDisplayName: providerName,
				Description:         appDesc,
				Labels:              appLabels,
				SystemAuthsIDs:      make([]string, 0),
				ApiBundles: []kymaModel.APIBundle{
					fixInternalAPIBundle("1"),
					fixInternalAPIBundle("2"),
					fixInternalAPIBundle("3"),
				},
			},
		},
		{
			description: "convert Compass App with empty Bundle pages",
			compassApp: Application{
				ID:           appId,
				Name:         appName,
				Description:  &appDesc,
				ProviderName: &providerName,
				Labels:       appLabels,
			},
			expectedApp: kymaModel.Application{
				ID:                  appId,
				Name:                appName,
				Description:         appDesc,
				ProviderDisplayName: providerName,
				Labels:              appLabels,
				SystemAuthsIDs:      make([]string, 0),
			},
		},
		{
			description: "convert Compass App with bundles using empty specs",
			compassApp: Application{
				ID:           appId,
				Name:         appName,
				Description:  &appDesc,
				ProviderName: &providerName,
				Bundles: &graphql.BundlePageExt{
					Data: []*graphql.BundleExt{
						fixCompassBundleExtWithEmptySpecs("1"),
						fixCompassBundleExtWithEmptySpecs("2"),
						fixCompassBundleExtWithEmptySpecs("3"),
					},
				},
			},
			expectedApp: kymaModel.Application{
				ID:                  appId,
				Name:                appName,
				Description:         appDesc,
				ProviderDisplayName: providerName,
				ApiBundles: []kymaModel.APIBundle{
					fixInternalAPIBundleEmptySpecs("1"),
					fixInternalAPIBundleEmptySpecs("2"),
					fixInternalAPIBundleEmptySpecs("3"),
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

func fixInternalAPIBundle(suffix string) kymaModel.APIBundle {
	return kymaModel.APIBundle{
		ID:                             baseBundleID + suffix,
		Name:                           baseBundleName + suffix,
		Description:                    stringPtr(baseBundleDesc + suffix),
		InstanceAuthRequestInputSchema: stringPtr(baseBundleInputSchema + suffix),
		APIDefinitions: []kymaModel.APIDefinition{
			fixInternalAPIDefinition("1", nil),
			fixInternalAPIDefinition("2", nil),
			fixInternalAPIDefinition("3", nil),
			fixInternalAPIDefinition("4", nil),
		},
		EventDefinitions: []kymaModel.EventAPIDefinition{
			fixInternalEventAPIDefinition("1"),
			fixInternalEventAPIDefinition("2"),
		},
	}
}

func fixInternalAPIBundleEmptySpecs(suffix string) kymaModel.APIBundle {
	return kymaModel.APIBundle{
		ID:                             baseBundleID + suffix,
		Name:                           baseBundleName + suffix,
		Description:                    stringPtr(baseBundleDesc + suffix),
		InstanceAuthRequestInputSchema: stringPtr(baseBundleInputSchema + suffix),
		APIDefinitions: []kymaModel.APIDefinition{
			fixInternalAPIDefinition("1", nil),
			fixInternalAPIDefinition("2", nil),
			fixInternalAPIDefinition("3", nil),
			fixInternalAPIDefinition("4", nil),
		},
		EventDefinitions: []kymaModel.EventAPIDefinition{
			fixInternalEventAPIDefinition("1"),
			fixInternalEventAPIDefinition("2"),
		},
	}
}

func fixInternalAPIDefinition(suffix string, credentials *kymaModel.Credentials) kymaModel.APIDefinition {
	return kymaModel.APIDefinition{
		ID:          baseAPIId + suffix,
		Name:        baseAPIName + suffix,
		Description: baseAPIDesc + suffix,
		TargetUrl:   baseAPIURL + suffix,
		Credentials: credentials,
	}
}

func fixInternalEventAPIDefinition(suffix string) kymaModel.EventAPIDefinition {
	return kymaModel.EventAPIDefinition{
		ID:          baseAPIId + suffix,
		Name:        baseAPIName + suffix,
		Description: baseAPIDesc + suffix,
	}
}

func fixCompassBundleWithDefaultInstanceAuth(suffix string, defaultInstanceAuth *graphql.Auth) *graphql.BundleExt {
	return &graphql.BundleExt{
		Bundle:           fixCompassBundle(suffix, defaultInstanceAuth),
		APIDefinitions:   fixAPIDefinitionPageExt(),
		EventDefinitions: fixEventAPIDefinitionPageExt(),
		Documents:        fixDocumentPageExt(),
	}
}

func fixInternalAPIBundleWithInstanceAuth(suffix string, defaultInstanceAuth *kymaModel.Auth) kymaModel.APIBundle {
	return kymaModel.APIBundle{
		ID:                             baseBundleID + suffix,
		Name:                           baseBundleName + suffix,
		Description:                    stringPtr(baseBundleDesc + suffix),
		InstanceAuthRequestInputSchema: stringPtr(baseBundleInputSchema + suffix),
		APIDefinitions: []kymaModel.APIDefinition{
			fixInternalAPIDefinition("1", nil),
			fixInternalAPIDefinition("2", nil),
			fixInternalAPIDefinition("3", nil),
			fixInternalAPIDefinition("4", nil),
		},
		EventDefinitions: []kymaModel.EventAPIDefinition{
			fixInternalEventAPIDefinition("1"),
			fixInternalEventAPIDefinition("2"),
		},
		DefaultInstanceAuth: defaultInstanceAuth,
	}
}

func fixCompassBundleExt(suffix string) *graphql.BundleExt {
	return &graphql.BundleExt{
		Bundle:           fixCompassBundle(suffix, nil),
		APIDefinitions:   fixAPIDefinitionPageExt(),
		EventDefinitions: fixEventAPIDefinitionPageExt(),
		Documents:        fixDocumentPageExt(),
	}
}

func fixCompassBundleExtWithEmptySpecs(suffix string) *graphql.BundleExt {
	return &graphql.BundleExt{
		Bundle:           fixCompassBundle(suffix, nil),
		APIDefinitions:   fixAPIDefinitionPageExt(),
		EventDefinitions: fixEventAPIDefinitionPageExtWithEmptySpecs(),
		Documents:        fixDocumentPageExtWithEmptyDocs(),
	}
}

func fixCompassBundle(suffix string, defaultInstanceAuth *graphql.Auth) graphql.Bundle {
	return graphql.Bundle{
		BaseEntity: &graphql.BaseEntity{
			ID: baseBundleID + suffix,
		},
		Name:                           baseBundleName + suffix,
		Description:                    stringPtr(baseBundleDesc + suffix),
		InstanceAuthRequestInputSchema: (*graphql.JSONSchema)(stringPtr(baseBundleInputSchema + suffix)),
		DefaultInstanceAuth:            defaultInstanceAuth,
	}
}

//func fixAPIDefinitionPageExt() graphql.APIDefinitionPageExt {
//	return graphql.APIDefinitionPageExt{
//		Data: []*graphql.APIDefinitionExt{
//			fixCompassAPIDefinitionExt("1", fixCompassOpenAPISpecExt()),
//			fixCompassAPIDefinitionExt("2", fixCompassOpenAPISpecExt()),
//			fixCompassAPIDefinitionExt("3", fixCompassODataSpecExt()),
//			fixCompassAPIDefinitionExt("4", fixCompassODataSpecExt()),
//		},
//	}
//}

func fixAPIDefinitionPageExt() graphql.APIDefinitionPageExt {
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
		BaseEntity: &graphql.BaseEntity{
			ID: baseAPIId + suffix,
		},
		Name:        baseAPIName + suffix,
		Description: &desc,
		TargetURL:   baseAPIURL + suffix,
		Spec:        spec,
	}
}

func fixCompassEventAPIDefinition(suffix string, spec *graphql.EventSpec) *graphql.EventDefinition {
	desc := baseAPIDesc + suffix

	return &graphql.EventDefinition{
		BaseEntity: &graphql.BaseEntity{
			ID: baseAPIId + suffix,
		},
		Name:        baseAPIName + suffix,
		Description: &desc,
		Spec:        spec,
	}
}

func fixCompassDocument(suffix string, data *graphql.CLOB) *graphql.Document {
	kind := baseDocKind + suffix

	return &graphql.Document{
		BaseEntity: &graphql.BaseEntity{
			ID: baseAPIId + suffix,
		},
		Description: baseAPIDesc + suffix,
		Title:       baseDocTitle + suffix,
		DisplayName: baseDocDisplayName + suffix,
		Format:      graphql.DocumentFormatMarkdown,
		Kind:        &kind,
		Data:        data,
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
