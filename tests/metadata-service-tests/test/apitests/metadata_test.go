package apitests

import (
	"net/http"
	"testing"

	"github.com/kyma-project/kyma/tests/metadata-service-tests/test/testkit"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

const (
	connectedApp = "connected-app"
)

func TestApiMetadata(t *testing.T) {

	config, err := testkit.ReadConfig()
	require.NoError(t, err)

	k8sResourcesClient, err := testkit.NewK8sInClusterResourcesClient(config.Namespace)
	require.NoError(t, err)

	dummyRE, err := k8sResourcesClient.CreateDummyRemoteEnvironment("dummy-re", v1.GetOptions{})
	require.NoError(t, err)

	metadataServiceClient := testkit.NewMetadataServiceClient(config.MetadataServiceUrl + "/" + dummyRE.Name + "/v1/metadata/services")

	expectedLabels := map[string]string{connectedApp: "dummy-re"}

	t.Run("registration API", func(t *testing.T) {
		t.Run("should register a service with OAuth credentials (with API, Events catalog, Documentation)", func(t *testing.T) {
			// when
			initialServiceDefinition := prepareOAuthServiceDetails("service-identifier", map[string]string{})

			statusCode, postResponseData, err := metadataServiceClient.CreateService(t, initialServiceDefinition)
			require.NoError(t, err)

			// then
			require.Equal(t, http.StatusOK, statusCode)
			require.NotEmpty(t, postResponseData.ID)

			// when
			statusCode, receivedServiceDefinition, err := metadataServiceClient.GetService(t, postResponseData.ID)
			require.NoError(t, err)
			expectedServiceDefinition := getExpectedDefinition(initialServiceDefinition, expectedLabels, receivedServiceDefinition.Identifier)

			// then
			require.Equal(t, http.StatusOK, statusCode)
			require.EqualValues(t, expectedServiceDefinition, *receivedServiceDefinition)

			// when
			statusCode, existingServices, err := metadataServiceClient.GetAllServices(t)
			require.NoError(t, err)

			postedService := findPostedService(existingServices, postResponseData.ID)

			// then
			require.Equal(t, http.StatusOK, statusCode)
			require.NotNil(t, postedService)
			require.Equal(t, "test service", postedService.Name)
			require.Equal(t, "service provider", postedService.Provider)
			require.Equal(t, "service description", postedService.Description)
			require.True(t, strings.HasPrefix(postedService.Identifier, "service-identifier"))
			require.Equal(t, map[string]string{connectedApp: "dummy-re"}, postedService.Labels)

			// clean up
			statusCode, err = metadataServiceClient.DeleteService(t, postResponseData.ID)
			require.NoError(t, err)

			require.Equal(t, http.StatusNoContent, statusCode)
		})

		t.Run("should register a service with Basic Auth credentials (with API, Events catalog, Documentation", func(t *testing.T) {
			// when
			initialServiceDefinition := prepareBasicAuthServiceDetails("service-identifier", map[string]string{})

			statusCode, postResponseData, err := metadataServiceClient.CreateService(t, initialServiceDefinition)
			require.NoError(t, err)

			// then
			require.Equal(t, http.StatusOK, statusCode)
			require.NotEmpty(t, postResponseData.ID)

			// when
			statusCode, receivedServiceDefinition, err := metadataServiceClient.GetService(t, postResponseData.ID)
			require.NoError(t, err)
			expectedServiceDefinition := getExpectedDefinition(initialServiceDefinition, expectedLabels, receivedServiceDefinition.Identifier)

			// then
			require.Equal(t, http.StatusOK, statusCode)
			require.EqualValues(t, expectedServiceDefinition, *receivedServiceDefinition)

			// when
			statusCode, existingServices, err := metadataServiceClient.GetAllServices(t)
			require.NoError(t, err)

			postedService := findPostedService(existingServices, postResponseData.ID)

			// then
			require.Equal(t, http.StatusOK, statusCode)
			require.NotNil(t, postedService)
			require.Equal(t, "test service", postedService.Name)
			require.Equal(t, "service provider", postedService.Provider)
			require.Equal(t, "service description", postedService.Description)
			require.True(t, strings.HasPrefix(postedService.Identifier, "service-identifier"))
			require.Equal(t, map[string]string{connectedApp: "dummy-re"}, postedService.Labels)

			// clean up
			statusCode, err = metadataServiceClient.DeleteService(t, postResponseData.ID)
			require.NoError(t, err)

			require.Equal(t, http.StatusNoContent, statusCode)
		})

		t.Run("should register service (overriding connected-app label)", func(t *testing.T) {
			// when
			initialServiceDefinition := prepareOAuthServiceDetails("service-identifier-2", map[string]string{"connected-app": "different-re"})

			statusCode, postResponseData, err := metadataServiceClient.CreateService(t, initialServiceDefinition)
			require.NoError(t, err)

			// then
			require.Equal(t, http.StatusOK, statusCode)
			require.NotEmpty(t, postResponseData.ID)

			// when
			statusCode, receivedServiceDefinition, err := metadataServiceClient.GetService(t, postResponseData.ID)
			require.NoError(t, err)
			expectedServiceDefinition := getExpectedDefinition(initialServiceDefinition, expectedLabels, receivedServiceDefinition.Identifier)

			// then
			require.Equal(t, http.StatusOK, statusCode)
			require.EqualValues(t, expectedServiceDefinition, *receivedServiceDefinition)

			// when
			statusCode, existingServices, err := metadataServiceClient.GetAllServices(t)
			require.NoError(t, err)

			postedService := findPostedService(existingServices, postResponseData.ID)

			// then
			require.Equal(t, http.StatusOK, statusCode)
			require.NotNil(t, postedService)
			require.Equal(t, "test service", postedService.Name)
			require.Equal(t, "service provider", postedService.Provider)
			require.Equal(t, "service description", postedService.Description)
			require.True(t, strings.HasPrefix(postedService.Identifier, "service-identifier-2"))
			require.Equal(t, map[string]string{connectedApp: "dummy-re"}, postedService.Labels)

			// clean up
			statusCode, err = metadataServiceClient.DeleteService(t, postResponseData.ID)
			require.NoError(t, err)
			initialServiceDefinition.Labels = map[string]string{}

			require.Equal(t, http.StatusNoContent, statusCode)
		})

		t.Run("should update service with OAuth credentials (with API, Events catalog, Documentation)", func(t *testing.T) {
			// given
			initialServiceDefinition := prepareOAuthServiceDetails("service-identifier-3", map[string]string{})

			postStatusCode, postResponseData, err := metadataServiceClient.CreateService(t, initialServiceDefinition)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, postStatusCode)

			updatedServiceDefinition := testkit.ServiceDetails{
				Name:        "updated test service",
				Provider:    "updated service provider",
				Description: "updated service description",
				Api: &testkit.API{
					TargetUrl: "http://service.com",
					Credentials: &testkit.Credentials{
						Oauth: &testkit.Oauth{
							URL:          "http://oauth.com",
							ClientID:     "clientId",
							ClientSecret: "clientSecret",
						},
					},
					Spec: testkit.ApiRawSpec,
				},
				Events: &testkit.Events{
					Spec: testkit.EventsRawSpec,
				},
				Documentation: &testkit.Documentation{
					DisplayName: "documentation name",
					Description: "documentation description",
					Type:        "documentation type",
					Tags:        []string{"tag1", "tag2"},
					Docs:        []testkit.DocsObject{{Title: "docs title", Type: "docs type", Source: "docs source"}},
				},
			}

			// when
			statusCode, err := metadataServiceClient.UpdateService(t, postResponseData.ID, updatedServiceDefinition)
			require.NoError(t, err)

			// then
			require.Equal(t, http.StatusOK, statusCode)

			// when
			statusCode, receivedServiceDefinition, err := metadataServiceClient.GetService(t, postResponseData.ID)
			require.NoError(t, err)
			expectedServiceDefinition := getExpectedDefinition(updatedServiceDefinition, expectedLabels, receivedServiceDefinition.Identifier)

			// then
			require.Equal(t, http.StatusOK, statusCode)
			require.EqualValues(t, expectedServiceDefinition, *receivedServiceDefinition)

			// when
			statusCode, existingServices, err := metadataServiceClient.GetAllServices(t)
			require.NoError(t, err)

			updatedService := findPostedService(existingServices, postResponseData.ID)

			// then
			require.Equal(t, http.StatusOK, statusCode)
			require.NotNil(t, updatedService)
			require.Equal(t, "updated test service", updatedService.Name)
			require.Equal(t, "updated service provider", updatedService.Provider)
			require.Equal(t, "updated service description", updatedService.Description)

			// clean up
			statusCode, err = metadataServiceClient.DeleteService(t, postResponseData.ID)
			require.NoError(t, err)

			require.Equal(t, http.StatusNoContent, statusCode)
		})

		t.Run("should update a service with Basic Auth credentials (with API, Events catalog, Documentation)", func(t *testing.T) {
			// given
			initialServiceDefinition := prepareOAuthServiceDetails("service-identifier-3", map[string]string{})

			postStatusCode, postResponseData, err := metadataServiceClient.CreateService(t, initialServiceDefinition)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, postStatusCode)

			updatedServiceDefinition := testkit.ServiceDetails{
				Name:        "updated test service",
				Provider:    "updated service provider",
				Description: "updated service description",
				Api: &testkit.API{
					TargetUrl: "http://service.com",
					Credentials: &testkit.Credentials{
						Basic: &testkit.Basic{
							Username: "username",
							Password: "password",
						},
					},
					Spec: testkit.ApiRawSpec,
				},
				Events: &testkit.Events{
					Spec: testkit.EventsRawSpec,
				},
				Documentation: &testkit.Documentation{
					DisplayName: "documentation name",
					Description: "documentation description",
					Type:        "documentation type",
					Tags:        []string{"tag1", "tag2"},
					Docs:        []testkit.DocsObject{{Title: "docs title", Type: "docs type", Source: "docs source"}},
				},
			}

			// when
			statusCode, err := metadataServiceClient.UpdateService(t, postResponseData.ID, updatedServiceDefinition)
			require.NoError(t, err)

			// then
			require.Equal(t, http.StatusOK, statusCode)

			// when
			statusCode, receivedServiceDefinition, err := metadataServiceClient.GetService(t, postResponseData.ID)
			require.NoError(t, err)
			expectedServiceDefinition := getExpectedDefinition(updatedServiceDefinition, expectedLabels, receivedServiceDefinition.Identifier)

			// then
			require.Equal(t, http.StatusOK, statusCode)
			require.EqualValues(t, expectedServiceDefinition, *receivedServiceDefinition)

			// when
			statusCode, existingServices, err := metadataServiceClient.GetAllServices(t)
			require.NoError(t, err)

			updatedService := findPostedService(existingServices, postResponseData.ID)

			// then
			require.Equal(t, http.StatusOK, statusCode)
			require.NotNil(t, updatedService)
			require.Equal(t, "updated test service", updatedService.Name)
			require.Equal(t, "updated service provider", updatedService.Provider)
			require.Equal(t, "updated service description", updatedService.Description)

			// clean up
			statusCode, err = metadataServiceClient.DeleteService(t, postResponseData.ID)
			require.NoError(t, err)

			require.Equal(t, http.StatusNoContent, statusCode)
		})

		t.Run("should return not found 404 when updating not existing service", func(t *testing.T) {
			// given
			updatedServiceDefinition := testkit.ServiceDetails{
				Name:        "updated test service",
				Provider:    "updated service provider",
				Description: "updated service description",
				Api: &testkit.API{
					TargetUrl: "http://service.com",
					Credentials: &testkit.Credentials{
						Oauth: &testkit.Oauth{
							URL:          "http://oauth.com",
							ClientID:     "clientId",
							ClientSecret: "clientSecret",
						},
					},
					Spec: testkit.ApiRawSpec,
				},
				Events: &testkit.Events{
					Spec: testkit.EventsRawSpec,
				},
				Documentation: &testkit.Documentation{
					DisplayName: "documentation name",
					Description: "documentation description",
					Type:        "documentation type",
					Tags:        []string{"tag1", "tag2"},
					Docs:        []testkit.DocsObject{{Title: "docs title", Type: "docs type", Source: "docs source"}},
				},
			}

			// when
			statusCode, err := metadataServiceClient.UpdateService(t, "12345", updatedServiceDefinition)

			// then
			require.NoError(t, err)
			require.Equal(t, http.StatusNotFound, statusCode)
		})

		t.Run("should delete service (with API, Events catalog, Documentation) - setup", func(t *testing.T) {
			// given
			initialServiceDefinition := prepareOAuthServiceDetails("service-identifier-4", map[string]string{})

			postStatusCode, postResponseData, err := metadataServiceClient.CreateService(t, initialServiceDefinition)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, postStatusCode)

			// when
			statusCode, err := metadataServiceClient.DeleteService(t, postResponseData.ID)

			// then
			require.Equal(t, http.StatusNoContent, statusCode)

			// when
			statusCode, _, err = metadataServiceClient.GetService(t, postResponseData.ID)
			require.NoError(t, err)

			// then
			require.Equal(t, http.StatusNotFound, statusCode)

			// when
			statusCode, existingServices, err := metadataServiceClient.GetAllServices(t)
			require.NoError(t, err)

			deletedService := findPostedService(existingServices, postResponseData.ID)

			// then
			require.Equal(t, http.StatusOK, statusCode)
			require.Nil(t, deletedService)
		})

		t.Run("should get service (with API, Events catalog, Documentation) - setup", func(t *testing.T) {
			// given
			initialServiceDefinition := prepareOAuthServiceDetails("service-identifier-5", map[string]string{})

			postStatusCode, postResponseData, err := metadataServiceClient.CreateService(t, initialServiceDefinition)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, postStatusCode)

			// when
			statusCode, receivedServiceDefinition, err := metadataServiceClient.GetService(t, postResponseData.ID)
			require.NoError(t, err)
			expectedServiceDefinition := getExpectedDefinition(initialServiceDefinition, expectedLabels, receivedServiceDefinition.Identifier)

			// then
			require.Equal(t, http.StatusOK, statusCode)
			require.EqualValues(t, expectedServiceDefinition, *receivedServiceDefinition)

			// when
			statusCode, existingServices, err := metadataServiceClient.GetAllServices(t)
			require.NoError(t, err)

			postedService := findPostedService(existingServices, postResponseData.ID)

			// then
			require.Equal(t, http.StatusOK, statusCode)
			require.NotNil(t, postedService)
			require.Equal(t, "test service", postedService.Name)
			require.Equal(t, "service provider", postedService.Provider)
			require.Equal(t, "service description", postedService.Description)

			// clean up
			statusCode, err = metadataServiceClient.DeleteService(t, postResponseData.ID)
			require.NoError(t, err)

			require.Equal(t, http.StatusNoContent, statusCode)
		})

		t.Run("should get all services (with API, Events catalog, Documentation)", func(t *testing.T) {
			// given
			initialServiceDefinition1 := prepareOAuthServiceDetails("service-identifier-6", map[string]string{})
			initialServiceDefinition2 := prepareOAuthServiceDetails("service-identifier-7", map[string]string{})

			postStatusCode1, postResponseData1, err := metadataServiceClient.CreateService(t, initialServiceDefinition1)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, postStatusCode1)

			postStatusCode2, postResponseData2, err := metadataServiceClient.CreateService(t, initialServiceDefinition2)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, postStatusCode2)

			// when
			statusCode, existingServices, err := metadataServiceClient.GetAllServices(t)
			require.NoError(t, err)

			// then
			require.Equal(t, http.StatusOK, statusCode)

			// when
			postedService1 := findPostedService(existingServices, postResponseData1.ID)
			postedService2 := findPostedService(existingServices, postResponseData2.ID)

			// then
			require.NotNil(t, postedService1)
			require.NotNil(t, postedService2)
			require.Equal(t, "test service", postedService1.Name)
			require.Equal(t, "service provider", postedService1.Provider)
			require.Equal(t, "service description", postedService1.Description)
			require.Equal(t, "test service", postedService2.Name)
			require.Equal(t, "service provider", postedService2.Provider)
			require.Equal(t, "service description", postedService2.Description)

			// clean up
			statusCode1, err := metadataServiceClient.DeleteService(t, postResponseData1.ID)
			require.NoError(t, err)

			statusCode2, err := metadataServiceClient.DeleteService(t, postResponseData2.ID)
			require.NoError(t, err)

			require.Equal(t, http.StatusNoContent, statusCode1)
			require.Equal(t, http.StatusNoContent, statusCode2)
		})

	})

	err = k8sResourcesClient.DeleteRemoteEnvironment(dummyRE.Name, &v1.DeleteOptions{})
	require.NoError(t, err)
}

func prepareOAuthServiceDetails(identifier string, labels map[string]string) testkit.ServiceDetails {
	return testkit.ServiceDetails{
		Name:        "test service",
		Provider:    "service provider",
		Description: "service description",
		Identifier:  identifier,
		Labels:      labels,
		Api: &testkit.API{
			TargetUrl: "http://service.com",
			Credentials: &testkit.Credentials{
				Oauth: &testkit.Oauth{
					URL:          "http://oauth.com",
					ClientID:     "clientId",
					ClientSecret: "clientSecret",
				},
			},
			Spec: testkit.ApiRawSpec,
		},
		Events: &testkit.Events{
			Spec: testkit.EventsRawSpec,
		},
		Documentation: &testkit.Documentation{
			DisplayName: "documentation name",
			Description: "documentation description",
			Type:        "documentation type",
			Tags:        []string{"tag1", "tag2"},
			Docs:        []testkit.DocsObject{{Title: "docs title", Type: "docs type", Source: "docs source"}},
		},
	}
}

func prepareBasicAuthServiceDetails(identifier string, labels map[string]string) testkit.ServiceDetails {
	return testkit.ServiceDetails{
		Name:        "test service",
		Provider:    "service provider",
		Description: "service description",
		Identifier:  identifier,
		Labels:      labels,
		Api: &testkit.API{
			TargetUrl: "http://service.com",
			Credentials: &testkit.Credentials{
				Basic: &testkit.Basic{
					Username: "username",
					Password: "password",
				},
			},
			Spec: testkit.ApiRawSpec,
		},
		Events: &testkit.Events{
			Spec: testkit.EventsRawSpec,
		},
		Documentation: &testkit.Documentation{
			DisplayName: "documentation name",
			Description: "documentation description",
			Type:        "documentation type",
			Tags:        []string{"tag1", "tag2"},
			Docs:        []testkit.DocsObject{{Title: "docs title", Type: "docs type", Source: "docs source"}},
		},
	}
}

func findPostedService(existingServices []testkit.Service, searchedID string) *testkit.Service {
	for _, e := range existingServices {
		if e.ID == searchedID {
			return &e
		}
	}
	return nil
}

func getExpectedDefinition(initialDefinition testkit.ServiceDetails, expectedLabels map[string]string, identifier string) testkit.ServiceDetails {
	initialDefinition.Labels = expectedLabels
	initialDefinition.Identifier = identifier
	return hideClientCredentials(initialDefinition)
}

func hideClientCredentials(original testkit.ServiceDetails) testkit.ServiceDetails {
	result := testkit.ServiceDetails{
		Name:        original.Name,
		Provider:    original.Provider,
		Description: original.Description,
		Identifier:  original.Identifier,
		Labels:      original.Labels,
	}

	if original.Api != nil {
		result.Api = &testkit.API{
			TargetUrl: original.Api.TargetUrl,
			Spec:      original.Api.Spec,
		}

		if original.Api.Credentials != nil {
			if original.Api.Credentials.Oauth != nil {
				result.Api.Credentials = &testkit.Credentials{
					Oauth: &testkit.Oauth{
						URL:          "http://oauth.com",
						ClientID:     "********",
						ClientSecret: "********",
					},
				}
			}

			if original.Api.Credentials.Basic != nil {
				result.Api.Credentials = &testkit.Credentials{
					Basic: &testkit.Basic{
						Username: "********",
						Password: "********",
					},
				}
			}
		}
	}

	if original.Events != nil {
		result.Events = &testkit.Events{
			Spec: original.Events.Spec,
		}
	}

	if original.Documentation != nil {
		result.Documentation = &testkit.Documentation{
			DisplayName: original.Documentation.DisplayName,
			Description: original.Documentation.Description,
			Type:        original.Documentation.Type,
			Tags:        []string{"tag1", "tag2"},
			Docs:        []testkit.DocsObject{{Title: "docs title", Type: "docs type", Source: "docs source"}},
		}

		if original.Documentation.Tags != nil {
			newTags := make([]string, len(original.Documentation.Tags))
			copy(newTags, original.Documentation.Tags)
			result.Documentation.Tags = newTags
		}

		if original.Documentation.Docs != nil {
			newDocs := make([]testkit.DocsObject, len(original.Documentation.Docs))
			copy(newDocs, original.Documentation.Docs)
			result.Documentation.Docs = newDocs
		}
	}

	return result
}
