package apitests

import (
	"net/http"
	"testing"

	"fmt"
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

	oauthAPI := &testkit.API{
		TargetUrl: "http://service.com",
		Credentials: &testkit.Credentials{
			Oauth: &testkit.Oauth{
				URL:          "http://oauth.com",
				ClientID:     "clientId",
				ClientSecret: "clientSecret",
			},
		},
		Spec: testkit.ApiRawSpec,
	}

	basicAuthAPI := &testkit.API{
		TargetUrl: "http://service.com",
		Credentials: &testkit.Credentials{
			Basic: &testkit.Basic{
				Username: "username",
				Password: "password",
			},
		},
		Spec: testkit.ApiRawSpec,
	}

	oauthAndBasicAuthAPI := &testkit.API{
		TargetUrl: "http://service.com",
		Credentials: &testkit.Credentials{
			Basic: &testkit.Basic{
				Username: "username",
				Password: "password",
			},
			Oauth: &testkit.Oauth{
				URL:          "http://oauth.com",
				ClientID:     "clientId",
				ClientSecret: "clientSecret",
			},
		},
		Spec: testkit.ApiRawSpec,
	}

	specAndSpecUrlAPI := &testkit.API{
		TargetUrl:        "http://service.com",
		Spec:             testkit.ApiRawSpec,
		SpecificationUrl: "http://some-url.com",
	}

	swaggerSpecAPI := &testkit.API{
		TargetUrl: "http://service.com",
		Spec:      testkit.SwaggerApiSpec,
	}

	t.Run("registration API", func(t *testing.T) {
		t.Run("should register a service with OAuth credentials (with API, Events catalog, Documentation)", func(t *testing.T) {
			// when
			initialServiceDefinition := prepareServiceDetails("service-identifier", map[string]string{}).WithAPI(oauthAPI)

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
			initialServiceDefinition := prepareServiceDetails("service-identifier-2", map[string]string{}).WithAPI(basicAuthAPI)

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

			require.Equal(t, http.StatusNoContent, statusCode)
		})

		t.Run("should return 400 when both OAuth and BasicAuth credentials provided", func(t *testing.T) {
			// when
			initialServiceDefinition := prepareServiceDetails("service-identifier-3", map[string]string{"connected-app": "dummy-re"}).WithAPI(oauthAndBasicAuthAPI)

			statusCode, _, err := metadataServiceClient.CreateService(t, initialServiceDefinition)
			require.NoError(t, err)

			// then
			require.Equal(t, http.StatusBadRequest, statusCode)
		})

		t.Run("should register service (overriding connected-app label)", func(t *testing.T) {
			// when
			initialServiceDefinition := prepareServiceDetails("service-identifier-4", map[string]string{"connected-app": "different-re"}).WithAPI(oauthAPI)

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
			require.True(t, strings.HasPrefix(postedService.Identifier, "service-identifier-4"))
			require.Equal(t, map[string]string{connectedApp: "dummy-re"}, postedService.Labels)

			// clean up
			statusCode, err = metadataServiceClient.DeleteService(t, postResponseData.ID)
			require.NoError(t, err)
			initialServiceDefinition.Labels = map[string]string{}

			require.Equal(t, http.StatusNoContent, statusCode)
		})

		t.Run("should register service modifying swagger api spec", func(t *testing.T) {
			// when
			initialServiceDefinition := prepareServiceDetails("service-identifier-5", map[string]string{"connected-app": "dummy-re"}).WithAPI(swaggerSpecAPI)

			statusCode, postResponseData, err := metadataServiceClient.CreateService(t, initialServiceDefinition)
			require.NoError(t, err)

			// then
			require.Equal(t, http.StatusOK, statusCode)
			require.NotEmpty(t, postResponseData.ID)

			// when
			statusCode, receivedServiceDefinition, err := metadataServiceClient.GetService(t, postResponseData.ID)
			require.NoError(t, err)
			expectedServiceDefinition := getExpectedDefinition(initialServiceDefinition, expectedLabels, receivedServiceDefinition.Identifier)
			expectedServiceDefinition.Api.Spec = modifiedSwaggerSpec(dummyRE.Name, postResponseData.ID, config.Namespace)

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
			require.True(t, strings.HasPrefix(postedService.Identifier, "service-identifier-5"))
			require.Equal(t, map[string]string{connectedApp: "dummy-re"}, postedService.Labels)

			// clean up
			statusCode, err = metadataServiceClient.DeleteService(t, postResponseData.ID)
			require.NoError(t, err)
			initialServiceDefinition.Labels = map[string]string{}

			require.Equal(t, http.StatusNoContent, statusCode)
		})

		t.Run("should register service using inline spec when both inline spec and spec url provided", func(t *testing.T) {
			// when
			initialServiceDefinition := prepareServiceDetails("service-identifier-6", map[string]string{"connected-app": "dummy-re"}).WithAPI(specAndSpecUrlAPI)

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
			require.True(t, strings.HasPrefix(postedService.Identifier, "service-identifier-6"))
			require.Equal(t, map[string]string{connectedApp: "dummy-re"}, postedService.Labels)

			// clean up
			statusCode, err = metadataServiceClient.DeleteService(t, postResponseData.ID)
			require.NoError(t, err)
			initialServiceDefinition.Labels = map[string]string{}

			require.Equal(t, http.StatusNoContent, statusCode)
		})

		t.Run("should update service with OAuth credentials (with API, Events catalog, Documentation)", func(t *testing.T) {
			// given
			initialServiceDefinition := prepareServiceDetails("service-identifier-8", map[string]string{}).WithAPI(oauthAPI)

			postStatusCode, postResponseData, err := metadataServiceClient.CreateService(t, initialServiceDefinition)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, postStatusCode)

			updatedServiceDefinition := testkit.ServiceDetails{
				Name:        "updated test service",
				Provider:    "updated service provider",
				Description: "updated service description",
				Api: &testkit.API{
					TargetUrl: "http://updated-service.com",
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
			require.True(t, strings.HasPrefix(updatedService.Identifier, "service-identifier-8"))

			// clean up
			statusCode, err = metadataServiceClient.DeleteService(t, postResponseData.ID)
			require.NoError(t, err)

			require.Equal(t, http.StatusNoContent, statusCode)
		})

		t.Run("should update a service with Basic Auth credentials (with API, Events catalog, Documentation)", func(t *testing.T) {
			// given
			initialServiceDefinition := prepareServiceDetails("service-identifier-9", map[string]string{}).WithAPI(oauthAPI)

			postStatusCode, postResponseData, err := metadataServiceClient.CreateService(t, initialServiceDefinition)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, postStatusCode)

			updatedServiceDefinition := testkit.ServiceDetails{
				Name:        "updated test service",
				Provider:    "updated service provider",
				Description: "updated service description",
				Api: &testkit.API{
					TargetUrl: "http://updated-service.com",
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
			require.True(t, strings.HasPrefix(updatedService.Identifier, "service-identifier-9"))

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
					TargetUrl: "http://updated-service.com",
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
			initialServiceDefinition := prepareServiceDetails("service-identifier-10", map[string]string{}).WithAPI(oauthAPI)

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
			initialServiceDefinition := prepareServiceDetails("service-identifier-11", map[string]string{}).WithAPI(oauthAPI)

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
			require.True(t, strings.HasPrefix(postedService.Identifier, "service-identifier-11"))

			// clean up
			statusCode, err = metadataServiceClient.DeleteService(t, postResponseData.ID)
			require.NoError(t, err)

			require.Equal(t, http.StatusNoContent, statusCode)
		})

		t.Run("should get all services (with API, Events catalog, Documentation)", func(t *testing.T) {
			// given
			initialServiceDefinition1 := prepareServiceDetails("service-identifier-12", map[string]string{}).WithAPI(oauthAPI)
			initialServiceDefinition2 := prepareServiceDetails("service-identifier-13", map[string]string{}).WithAPI(oauthAPI)

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
			require.True(t, strings.HasPrefix(postedService1.Identifier, "service-identifier-12"))
			require.Equal(t, "test service", postedService2.Name)
			require.Equal(t, "service provider", postedService2.Provider)
			require.Equal(t, "service description", postedService2.Description)
			require.True(t, strings.HasPrefix(postedService2.Identifier, "service-identifier-13"))

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

func prepareServiceDetails(identifier string, labels map[string]string) testkit.ServiceDetails {
	return testkit.ServiceDetails{
		Name:        "test service",
		Provider:    "service provider",
		Description: "service description",
		Identifier:  identifier,
		Labels:      labels,
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

func modifiedSwaggerSpec(reName string, serviceId string, namespace string) []byte {
	return testkit.Compact([]byte(
		fmt.Sprintf("{\"schemes\":[\"http\"],\"swagger\":\"2.0\",\"host\":\"re-%s-%s.%s.svc.cluster.local\",\"paths\":null}", reName, serviceId, namespace)),
	)
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
			TargetUrl:        original.Api.TargetUrl,
			Spec:             original.Api.Spec,
			SpecificationUrl: original.Api.SpecificationUrl,
			ApiType:          original.Api.ApiType,
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
