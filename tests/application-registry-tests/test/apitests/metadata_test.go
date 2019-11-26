package apitests

import (
	"net/http"
	"testing"

	"fmt"
	"strings"

	"github.com/kyma-project/kyma/tests/application-registry-tests/test/testkit"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	connectedApp         = "connected-app"
	csrfTokenEndpointURL = "https://csrf.token.endpoint.org"
)

var (
	headers = map[string][]string{
		"headerKey": {"headerValue"},
	}
	queryParameters = map[string][]string{
		"queryParameterKey": {"queryParameterValue"},
	}
	requestParameters = testkit.RequestParameters{
		Headers:         &headers,
		QueryParameters: &queryParameters,
	}
)

func TestApiMetadata(t *testing.T) {

	config, err := testkit.ReadConfig()
	require.NoError(t, err)

	k8sResourcesClient, err := testkit.NewK8sInClusterResourcesClient(config.Namespace)
	require.NoError(t, err)

	dummyApp, err := k8sResourcesClient.CreateDummyApplication("appmetadatatest0", v1.GetOptions{}, true)
	require.NoError(t, err)

	metadataServiceClient := testkit.NewMetadataServiceClient(config.MetadataServiceUrl + "/" + dummyApp.Name + "/v1/metadata/services")

	expectedLabels := map[string]string{connectedApp: dummyApp.Name}

	oauthAPI := newOauthAPI()
	oauthAPIWithCSRF := newOauthAPI()
	oauthAPIWithCSRF.Credentials.Oauth.CSRFInfo = &testkit.CSRFInfo{TokenEndpointURL: csrfTokenEndpointURL}
	oauthAPIWithRequestParameters := newOauthAPI()
	oauthAPIWithRequestParameters.Credentials.Oauth.RequestParameters = &requestParameters

	basicAuthAPI := newBasicAuthAPI()
	basicAuthAPIWithCSRF := newBasicAuthAPI()
	basicAuthAPIWithCSRF.Credentials.Basic.CSRFInfo = &testkit.CSRFInfo{TokenEndpointURL: csrfTokenEndpointURL}

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

	certGenAPI := newCertGenAPI()
	certGenAPIWithCSRF := newCertGenAPI()
	certGenAPIWithCSRF.Credentials.CertificateGen.CSRFInfo = &testkit.CSRFInfo{TokenEndpointURL: csrfTokenEndpointURL}

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

		testOAuthAPI := func(api *testkit.API, t *testing.T) {
			// when
			identifier := testkit.GenerateIdentifier()
			initialServiceDefinition := prepareServiceDetails(identifier, map[string]string{}).WithAPI(api)

			statusCode, postResponseData, err := metadataServiceClient.CreateService(t, initialServiceDefinition)
			require.NoError(t, err)

			defer deleteService(t, metadataServiceClient, postResponseData)

			// then
			require.Equal(t, http.StatusOK, statusCode)
			require.NotEmpty(t, postResponseData.ID)

			// when
			statusCode, receivedServiceDefinition, err := metadataServiceClient.GetService(t, postResponseData.ID)
			require.NoError(t, err)
			expectedServiceDefinition := getExpectedDefinition(initialServiceDefinition, expectedLabels, receivedServiceDefinition.Identifier)

			// then
			require.Equal(t, http.StatusOK, statusCode)
			requireServiceDefinitionEqual(t, expectedServiceDefinition, receivedServiceDefinition)

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
			require.True(t, strings.HasPrefix(postedService.Identifier, identifier))
			require.Equal(t, map[string]string{connectedApp: dummyApp.Name}, postedService.Labels)
		}

		t.Run("should register a service with OAuth credentials (with API, Events catalog, Documentation)", func(t *testing.T) {
			testOAuthAPI(oauthAPI, t)
		})

		t.Run("should register a service with OAuth credentials and CSRF token (with API, Events catalog, Documentation)", func(t *testing.T) {
			testOAuthAPI(oauthAPIWithCSRF, t)
		})

		t.Run("should register a service with OAuth credentials and additional headers and query parameters (with API, Events catalog, Documentation)", func(t *testing.T) {
			testOAuthAPI(oauthAPIWithRequestParameters, t)
		})

		testBasicApi := func(api *testkit.API, t *testing.T) {
			// when
			identifier := testkit.GenerateIdentifier()
			initialServiceDefinition := prepareServiceDetails(identifier, map[string]string{}).WithAPI(api)

			statusCode, postResponseData, err := metadataServiceClient.CreateService(t, initialServiceDefinition)
			require.NoError(t, err)

			defer deleteService(t, metadataServiceClient, postResponseData)

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
			require.True(t, strings.HasPrefix(postedService.Identifier, identifier))
			require.Equal(t, map[string]string{connectedApp: dummyApp.Name}, postedService.Labels)
		}

		t.Run("should register a service with Basic Auth credentials (with API, Events catalog, Documentation", func(t *testing.T) {
			testBasicApi(basicAuthAPI, t)
		})

		t.Run("should register a service with Basic Auth credentials and CSRF token (with API, Events catalog, Documentation", func(t *testing.T) {
			testBasicApi(basicAuthAPIWithCSRF, t)
		})

		testCertGenApi := func(api *testkit.API, t *testing.T) {
			// when
			identifier := testkit.GenerateIdentifier()
			initialServiceDefinition := prepareServiceDetails(identifier, map[string]string{}).WithAPI(api)

			statusCode, postResponseData, err := metadataServiceClient.CreateService(t, initialServiceDefinition)
			require.NoError(t, err)

			defer deleteService(t, metadataServiceClient, postResponseData)

			// then
			require.Equal(t, http.StatusOK, statusCode)
			require.NotEmpty(t, postResponseData.ID)

			// when
			statusCode, receivedServiceDefinition, err := metadataServiceClient.GetService(t, postResponseData.ID)
			require.NoError(t, err)
			expectedServiceDefinition := getExpectedDefinition(initialServiceDefinition, expectedLabels, receivedServiceDefinition.Identifier)

			// then
			require.Equal(t, http.StatusOK, statusCode)
			requireDefinitionsWithCertCredentialsEqual(t, expectedServiceDefinition, *receivedServiceDefinition)
			requireCertificateNotEmpty(t, *receivedServiceDefinition)

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
			require.True(t, strings.HasPrefix(postedService.Identifier, identifier))
			require.Equal(t, map[string]string{connectedApp: dummyApp.Name}, postedService.Labels)
		}

		t.Run("should register a service with CertificateGen credentials (with API, Events catalog, Documentation", func(t *testing.T) {
			testCertGenApi(certGenAPI, t)
		})

		t.Run("should register a service with CertificateGen credentials and CSRF token (with API, Events catalog, Documentation", func(t *testing.T) {
			testCertGenApi(certGenAPIWithCSRF, t)
		})

		t.Run("should return 400 when both OAuth and BasicAuth credentials provided", func(t *testing.T) {
			// when
			identifier := testkit.GenerateIdentifier()
			initialServiceDefinition := prepareServiceDetails(identifier, map[string]string{"connected-app": dummyApp.Name}).WithAPI(oauthAndBasicAuthAPI)

			statusCode, _, err := metadataServiceClient.CreateService(t, initialServiceDefinition)
			require.NoError(t, err)

			// then
			require.Equal(t, http.StatusBadRequest, statusCode)
		})

		t.Run("should register service (overriding connected-app label)", func(t *testing.T) {
			// when
			identifier := testkit.GenerateIdentifier()
			initialServiceDefinition := prepareServiceDetails(identifier, map[string]string{"connected-app": "different-re"}).WithAPI(oauthAPI)

			statusCode, postResponseData, err := metadataServiceClient.CreateService(t, initialServiceDefinition)
			require.NoError(t, err)

			defer deleteService(t, metadataServiceClient, postResponseData)

			// then
			require.Equal(t, http.StatusOK, statusCode)
			require.NotEmpty(t, postResponseData.ID)

			// when
			statusCode, receivedServiceDefinition, err := metadataServiceClient.GetService(t, postResponseData.ID)
			require.NoError(t, err)
			expectedServiceDefinition := getExpectedDefinition(initialServiceDefinition, expectedLabels, receivedServiceDefinition.Identifier)

			// then
			require.Equal(t, http.StatusOK, statusCode)
			requireServiceDefinitionEqual(t, expectedServiceDefinition, receivedServiceDefinition)

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
			require.True(t, strings.HasPrefix(postedService.Identifier, identifier))
			require.Equal(t, map[string]string{connectedApp: dummyApp.Name}, postedService.Labels)
		})

		t.Run("should register service adding connected-app label to the existing", func(t *testing.T) {
			// when
			identifier := testkit.GenerateIdentifier()
			initialServiceDefinition := prepareServiceDetails(identifier, map[string]string{"test": "test"}).WithAPI(oauthAPI)

			statusCode, postResponseData, err := metadataServiceClient.CreateService(t, initialServiceDefinition)
			require.NoError(t, err)

			defer func() {
				// clean up
				statusCode, err = metadataServiceClient.DeleteService(t, postResponseData.ID)
				require.NoError(t, err)
				require.Equal(t, http.StatusNoContent, statusCode)
			}()

			// then
			require.Equal(t, http.StatusOK, statusCode)
			require.NotEmpty(t, postResponseData.ID)

			// when
			statusCode, receivedServiceDefinition, err := metadataServiceClient.GetService(t, postResponseData.ID)
			require.NoError(t, err)
			expectedLabels := map[string]string{"test": "test", connectedApp: dummyApp.Name}
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
			require.True(t, strings.HasPrefix(postedService.Identifier, identifier))
			require.Equal(t, expectedLabels, postedService.Labels)
		})

		t.Run("should register service modifying swagger api spec", func(t *testing.T) {
			// when
			identifier := testkit.GenerateIdentifier()
			initialServiceDefinition := prepareServiceDetails(identifier, map[string]string{"connected-app": dummyApp.Name}).WithAPI(swaggerSpecAPI)

			statusCode, postResponseData, err := metadataServiceClient.CreateService(t, initialServiceDefinition)
			require.NoError(t, err)

			defer deleteService(t, metadataServiceClient, postResponseData)

			// then
			require.Equal(t, http.StatusOK, statusCode)
			require.NotEmpty(t, postResponseData.ID)

			// when
			statusCode, receivedServiceDefinition, err := metadataServiceClient.GetService(t, postResponseData.ID)
			require.NoError(t, err)
			expectedServiceDefinition := getExpectedDefinition(initialServiceDefinition, expectedLabels, receivedServiceDefinition.Identifier)
			expectedServiceDefinition.Api.Spec = modifiedSwaggerSpec(dummyApp.Name, postResponseData.ID, config.Namespace)

			// then
			require.Equal(t, http.StatusOK, statusCode)
			requireServiceDefinitionEqual(t, expectedServiceDefinition, receivedServiceDefinition)

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
			require.True(t, strings.HasPrefix(postedService.Identifier, identifier))
			require.Equal(t, map[string]string{connectedApp: dummyApp.Name}, postedService.Labels)
		})

		t.Run("should register service using inline spec when both inline spec and spec url provided", func(t *testing.T) {
			// when
			identifier := testkit.GenerateIdentifier()
			initialServiceDefinition := prepareServiceDetails(identifier, map[string]string{"connected-app": dummyApp.Name}).WithAPI(specAndSpecUrlAPI)

			statusCode, postResponseData, err := metadataServiceClient.CreateService(t, initialServiceDefinition)
			require.NoError(t, err)

			defer deleteService(t, metadataServiceClient, postResponseData)
			// then
			require.Equal(t, http.StatusOK, statusCode)
			require.NotEmpty(t, postResponseData.ID)

			// when
			statusCode, receivedServiceDefinition, err := metadataServiceClient.GetService(t, postResponseData.ID)
			require.NoError(t, err)
			expectedServiceDefinition := getExpectedDefinition(initialServiceDefinition, expectedLabels, receivedServiceDefinition.Identifier)

			// then
			require.Equal(t, http.StatusOK, statusCode)
			requireServiceDefinitionEqual(t, expectedServiceDefinition, receivedServiceDefinition)

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
			require.True(t, strings.HasPrefix(postedService.Identifier, identifier))
			require.Equal(t, map[string]string{connectedApp: dummyApp.Name}, postedService.Labels)
		})

		t.Run("should update service with OAuth credentials (with API, Events catalog, Documentation)", func(t *testing.T) {
			// given
			identifier := testkit.GenerateIdentifier()
			initialServiceDefinition := prepareServiceDetails(identifier, map[string]string{}).WithAPI(oauthAPI)

			postStatusCode, postResponseData, err := metadataServiceClient.CreateService(t, initialServiceDefinition)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, postStatusCode)

			defer deleteService(t, metadataServiceClient, postResponseData)

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
			requireServiceDefinitionEqual(t, expectedServiceDefinition, receivedServiceDefinition)

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
			require.True(t, strings.HasPrefix(updatedService.Identifier, identifier))
		})

		t.Run("should update a service with Basic Auth credentials (with API, Events catalog, Documentation)", func(t *testing.T) {
			// given
			identifier := testkit.GenerateIdentifier()
			initialServiceDefinition := prepareServiceDetails(identifier, map[string]string{}).WithAPI(oauthAPI)

			postStatusCode, postResponseData, err := metadataServiceClient.CreateService(t, initialServiceDefinition)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, postStatusCode)

			defer deleteService(t, metadataServiceClient, postResponseData)

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
			requireServiceDefinitionEqual(t, expectedServiceDefinition, receivedServiceDefinition)

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
			require.True(t, strings.HasPrefix(updatedService.Identifier, identifier))
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
			identifier := testkit.GenerateIdentifier()
			initialServiceDefinition := prepareServiceDetails(identifier, map[string]string{}).WithAPI(oauthAPI)

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
			identifier := testkit.GenerateIdentifier()
			initialServiceDefinition := prepareServiceDetails(identifier, map[string]string{}).WithAPI(oauthAPI)

			postStatusCode, postResponseData, err := metadataServiceClient.CreateService(t, initialServiceDefinition)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, postStatusCode)

			defer deleteService(t, metadataServiceClient, postResponseData)

			// when
			statusCode, receivedServiceDefinition, err := metadataServiceClient.GetService(t, postResponseData.ID)
			require.NoError(t, err)
			expectedServiceDefinition := getExpectedDefinition(initialServiceDefinition, expectedLabels, receivedServiceDefinition.Identifier)

			// then
			require.Equal(t, http.StatusOK, statusCode)
			requireServiceDefinitionEqual(t, expectedServiceDefinition, receivedServiceDefinition)

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
			require.True(t, strings.HasPrefix(postedService.Identifier, identifier))
		})

		t.Run("should get all services (with API, Events catalog, Documentation)", func(t *testing.T) {
			// given
			identifier1 := testkit.GenerateIdentifier()
			initialServiceDefinition1 := prepareServiceDetails(identifier1, map[string]string{}).WithAPI(oauthAPI)
			identifier2 := testkit.GenerateIdentifier()
			initialServiceDefinition2 := prepareServiceDetails(identifier2, map[string]string{}).WithAPI(oauthAPI)

			postStatusCode1, postResponseData1, err := metadataServiceClient.CreateService(t, initialServiceDefinition1)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, postStatusCode1)

			postStatusCode2, postResponseData2, err := metadataServiceClient.CreateService(t, initialServiceDefinition2)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, postStatusCode2)

			defer func() {
				// clean up
				statusCode1, err := metadataServiceClient.DeleteService(t, postResponseData1.ID)
				require.NoError(t, err)
				statusCode2, err := metadataServiceClient.DeleteService(t, postResponseData2.ID)
				require.NoError(t, err)
				require.Equal(t, http.StatusNoContent, statusCode1)
				require.Equal(t, http.StatusNoContent, statusCode2)
			}()

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
			require.True(t, strings.HasPrefix(postedService1.Identifier, identifier1))
			require.Equal(t, "test service", postedService2.Name)
			require.Equal(t, "service provider", postedService2.Provider)
			require.Equal(t, "service description", postedService2.Description)
			require.True(t, strings.HasPrefix(postedService2.Identifier, identifier2))
		})

	})

	err = k8sResourcesClient.DeleteApplication(dummyApp.Name, &v1.DeleteOptions{})
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

func modifiedSwaggerSpec(appName string, serviceId string, namespace string) []byte {
	return testkit.Compact([]byte(
		fmt.Sprintf("{\"schemes\":[\"http\"],\"swagger\":\"2.0\",\"host\":\"%s-%s.%s.svc.cluster.local\",\"paths\":null}", appName, serviceId, namespace)),
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

				var csrfInfo *testkit.CSRFInfo = nil
				if original.Api.Credentials.Oauth.CSRFInfo != nil {
					csrfInfo = original.Api.Credentials.Oauth.CSRFInfo
				}

				result.Api.Credentials = &testkit.Credentials{
					Oauth: &testkit.Oauth{
						URL:               "http://oauth.com",
						ClientID:          "********",
						ClientSecret:      "********",
						CSRFInfo:          csrfInfo,
						RequestParameters: original.Api.Credentials.Oauth.RequestParameters,
					},
				}
			}

			if original.Api.Credentials.Basic != nil {

				var csrfInfo *testkit.CSRFInfo = nil
				if original.Api.Credentials.Basic.CSRFInfo != nil {
					csrfInfo = original.Api.Credentials.Basic.CSRFInfo
				}

				result.Api.Credentials = &testkit.Credentials{
					Basic: &testkit.Basic{
						Username: "********",
						Password: "********",
						CSRFInfo: csrfInfo,
					},
				}
			}

			if original.Api.Credentials.CertificateGen != nil {

				result.Api.Credentials = &testkit.Credentials{
					CertificateGen: original.Api.Credentials.CertificateGen,
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

func requireDefinitionsWithCertCredentialsEqual(t *testing.T, expected testkit.ServiceDetails, actual testkit.ServiceDetails) {
	require.Equal(t, expected.Description, actual.Description)
	require.Equal(t, expected.Identifier, actual.Identifier)
	require.Equal(t, expected.Api.Credentials.CertificateGen.CommonName, actual.Api.Credentials.CertificateGen.CommonName)
	require.Equal(t, expected.Name, actual.Name)
	require.Equal(t, expected.Provider, actual.Provider)
	require.EqualValues(t, expected.Labels, actual.Labels)
	require.NotNil(t, actual.Events)
	require.EqualValues(t, expected.Documentation, actual.Documentation)
}

func requireServiceDefinitionEqual(t *testing.T, expected testkit.ServiceDetails, actual *testkit.ServiceDetails) {
	require.NotNil(t, actual)
	require.Equal(t, expected.Name, actual.Name)
	require.Equal(t, expected.Provider, actual.Provider)
	require.Equal(t, expected.Description, actual.Description)
	require.Equal(t, expected.ShortDescription, actual.ShortDescription)
	require.Equal(t, expected.Identifier, actual.Identifier)
	require.EqualValues(t, expected.Labels, actual.Labels)
	require.EqualValues(t, expected.Api, actual.Api)
	if expected.Events != nil {
		require.NotNil(t, actual.Events)
	} else {
		require.Nil(t, actual.Events)
	}
	require.EqualValues(t, expected.Documentation, actual.Documentation)
}

func requireCertificateNotEmpty(t *testing.T, actual testkit.ServiceDetails) {
	require.NotEmpty(t, actual.Api.Credentials.CertificateGen.Certificate)
}

func newOauthAPI() *testkit.API {
	return &testkit.API{
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
}

func newBasicAuthAPI() *testkit.API {
	return &testkit.API{
		TargetUrl: "http://service.com",
		Credentials: &testkit.Credentials{
			Basic: &testkit.Basic{
				Username: "username",
				Password: "password",
			},
		},
		Spec: testkit.ApiRawSpec,
	}
}

func newCertGenAPI() *testkit.API {
	return &testkit.API{
		TargetUrl: "http://service.com",
		Credentials: &testkit.Credentials{
			CertificateGen: &testkit.CertificateGen{
				CommonName: "commonName",
			},
		},
		Spec: testkit.ApiRawSpec,
	}
}

func deleteService(t *testing.T, metadataServiceClient testkit.MetadataServiceClient, postResponse *testkit.PostServiceResponse) {
	if postResponse != nil {
		statusCode, err := metadataServiceClient.DeleteService(t, postResponse.ID)
		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, statusCode)
	}
}
