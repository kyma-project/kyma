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
			description: "convert Compass App with bundles",
			compassApp: Application{
				ID:           appId,
				Name:         appName,
				Description:  &appDesc,
				ProviderName: &providerName,
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
				Description:         appDesc,
				ProviderDisplayName: providerName,
				ApiBundles: []kymaModel.APIBundle{
					fixInternalAPIBundle("1"),
					fixInternalAPIBundle("2"),
					fixInternalAPIBundle("3"),
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

func fixAPIDefinitionPageExt() graphql.APIDefinitionPageExt {
	return graphql.APIDefinitionPageExt{
		Data: []*graphql.APIDefinitionExt{
			fixCompassAPIDefinitionExt("1"),
			fixCompassAPIDefinitionExt("2"),
			fixCompassAPIDefinitionExt("3"),
			fixCompassAPIDefinitionExt("4"),
		},
	}
}

func fixCompassAPIDefinitionExt(suffix string) *graphql.APIDefinitionExt {
	apiDefinition := fixCompassAPIDefinition(suffix)

	return &graphql.APIDefinitionExt{
		APIDefinition: *apiDefinition,
	}
}

func fixEventAPIDefinitionPageExt() graphql.EventAPIDefinitionPageExt {
	return graphql.EventAPIDefinitionPageExt{
		Data: []*graphql.EventAPIDefinitionExt{
			fixEventAPIDefinitionExt("1"),
			fixEventAPIDefinitionExt("2"),
		},
	}
}

func fixEventAPIDefinitionExt(suffix string) *graphql.EventAPIDefinitionExt {
	eventDefinition := fixCompassEventAPIDefinition(suffix)

	return &graphql.EventAPIDefinitionExt{
		EventDefinition: *eventDefinition,
	}
}

func fixCompassAPIDefinition(suffix string) *graphql.APIDefinition {
	desc := baseAPIDesc + suffix

	return &graphql.APIDefinition{
		BaseEntity: &graphql.BaseEntity{
			ID: baseAPIId + suffix,
		},
		Name:        baseAPIName + suffix,
		Description: &desc,
		TargetURL:   baseAPIURL + suffix,
	}
}

func fixCompassEventAPIDefinition(suffix string) *graphql.EventDefinition {
	desc := baseAPIDesc + suffix

	return &graphql.EventDefinition{
		BaseEntity: &graphql.BaseEntity{
			ID: baseAPIId + suffix,
		},
		Name:        baseAPIName + suffix,
		Description: &desc,
	}
}

func stringPtr(str string) *string {
	return &str
}
