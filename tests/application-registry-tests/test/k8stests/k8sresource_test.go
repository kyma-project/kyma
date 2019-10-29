package k8stests

import (
	"time"

	"github.com/stretchr/testify/require"
	v1core "k8s.io/api/core/v1"

	"net/http"
	"testing"

	"github.com/kyma-project/kyma/tests/application-registry-tests/test/testkit"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	crPropagationWaitTime              = 10
	deleteApplicationResourcesWaitTime = 10
)

var (
	headers = map[string][]string{
		"headerKey": []string{"headerValue"},
	}
	queryParameters = map[string][]string{
		"queryParameterKey": []string{"queryParameterValue"},
	}
	requestParameters = testkit.RequestParameters{
		Headers:         &headers,
		QueryParameters: &queryParameters,
	}
)

func TestK8sResources(t *testing.T) {

	config, err := testkit.ReadConfig()
	require.NoError(t, err)

	k8sResourcesClient, err := testkit.NewK8sInClusterResourcesClient(config.Namespace)
	require.NoError(t, err)

	dummyApp, err := k8sResourcesClient.CreateDummyApplication("appk8srestest0", v1.GetOptions{}, true)
	require.NoError(t, err)

	defer k8sResourcesClient.DeleteApplication(dummyApp.Name, &v1.DeleteOptions{})

	metadataServiceClient := testkit.NewMetadataServiceClient(config.MetadataServiceUrl + "/" + dummyApp.Name + "/v1/metadata/services")

	defaultOAuthAPI := testkit.API{
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

	testWithOAuth := func(api *testkit.API, t *testing.T) {

		// setup
		initialServiceDefinition := testkit.ServiceDetails{
			Name:        "test service",
			Provider:    "service provider",
			Description: "service description",
			Api:         api,
		}

		statusCode, postResponseData, err := metadataServiceClient.CreateService(t, initialServiceDefinition)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		serviceId := postResponseData.ID
		defer metadataServiceClient.DeleteService(t, serviceId)

		resourceName := dummyApp.Name + "-" + serviceId
		paramsSecretName := testkit.CreateParamsSecretName(resourceName)

		expectedLabels := map[string]string{"app": dummyApp.Name, "serviceId": serviceId}

		time.Sleep(crPropagationWaitTime * time.Second)

		// tests
		t.Run("should create k8s service", func(t *testing.T) {
			k8sService, err := k8sResourcesClient.GetService(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sService(t, k8sService, resourceName, expectedLabels, v1core.ProtocolTCP, 80, 8080)
		})

		t.Run("should create k8s secret with client credentials", func(t *testing.T) {
			k8sSecret, err := k8sResourcesClient.GetSecret(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sOAuthSecret(t, k8sSecret, resourceName, expectedLabels, "clientId", "clientSecret")
		})

		t.Run("should create k8s secret with request parameters", func(t *testing.T) {
			k8sSecret, err := k8sResourcesClient.GetSecret(paramsSecretName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sParamsSecret(t, k8sSecret, paramsSecretName, expectedLabels, "headerKey", "headerValue", "queryParameterKey", "queryParameterValue")
		})

		t.Run("should create istio denier", func(t *testing.T) {
			denier, err := k8sResourcesClient.GetHandler(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sIstioHandler(t, denier, resourceName, expectedLabels, 7, "Not allowed")
		})

		t.Run("should create istio rule", func(t *testing.T) {
			rule, err := k8sResourcesClient.GetRule(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sIstioRule(t, rule, resourceName, config.Namespace, expectedLabels)
		})

		t.Run("should create istio checknothing", func(t *testing.T) {
			checknothing, err := k8sResourcesClient.GetInstance(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sIstioInstance(t, checknothing, resourceName, expectedLabels)
		})

		t.Run("should add service to application custom resource", func(t *testing.T) {
			application, err := k8sResourcesClient.GetApplicationServices(dummyApp.Name, v1.GetOptions{})
			require.NoError(t, err)

			expectedCSRF := "not used in this test run"
			if api.Credentials.Oauth.CSRFInfo != nil {
				expectedCSRF = api.Credentials.Oauth.CSRFInfo.TokenEndpointURL
			}

			expectedServiceData := testkit.ServiceData{
				ServiceId:            serviceId,
				DisplayName:          "test service",
				ProviderDisplayName:  "service provider",
				LongDescription:      "service description",
				HasAPI:               true,
				TargetUrl:            "http://service.com",
				OauthUrl:             "http://oauth.com",
				GatewayUrl:           "http://" + resourceName + ".kyma-integration.svc.cluster.local",
				AccessLabel:          resourceName,
				HasEvents:            false,
				CSRFTokenEndpointURL: expectedCSRF,
			}

			testkit.CheckK8sApplication(t, application, dummyApp.Name, expectedServiceData)
		})
	}

	t.Run("when creating service only with OAuth API", func(t *testing.T) {
		t.Run("when current api with request parameters provided", func(t *testing.T) {
			testWithOAuth(defaultOAuthAPI.WithRequestParameters(&requestParameters), t)
		})

		t.Run("when deprecated api provided", func(t *testing.T) {
			testWithOAuth(defaultOAuthAPI.WithHeadersAndQueryParameters(&headers, &queryParameters), t)
		})
	})

	t.Run("when creating service only with OAuth API and CSRF", func(t *testing.T) {
		csrfInfo := &testkit.CSRFInfo{TokenEndpointURL: "https://csrf.token.endpoint.org"}

		t.Run("when current api with request parameters provided", func(t *testing.T) {
			testWithOAuth(defaultOAuthAPI.WithRequestParameters(&requestParameters).WithCSRFInOAuth(csrfInfo), t)
		})

		t.Run("when deprecated api provided", func(t *testing.T) {
			testWithOAuth(defaultOAuthAPI.WithHeadersAndQueryParameters(&headers, &queryParameters).WithCSRFInOAuth(csrfInfo), t)
		})
	})

	defaultBasicAPI := testkit.API{
		TargetUrl: "http://service.com",
		Credentials: &testkit.Credentials{
			Basic: &testkit.Basic{
				Username: "username",
				Password: "password",
			},
		},
		Spec: testkit.ApiRawSpec,
	}

	testWithBasicAuth := func(api *testkit.API, t *testing.T) {

		// setup
		initialServiceDefinition := testkit.ServiceDetails{
			Name:        "test service",
			Provider:    "service provider",
			Description: "service description",
			Api:         api,
		}

		statusCode, postResponseData, err := metadataServiceClient.CreateService(t, initialServiceDefinition)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		serviceId := postResponseData.ID
		defer metadataServiceClient.DeleteService(t, serviceId)

		resourceName := dummyApp.Name + "-" + serviceId
		paramsSecretName := testkit.CreateParamsSecretName(resourceName)

		expectedLabels := map[string]string{"app": dummyApp.Name, "serviceId": serviceId}

		time.Sleep(crPropagationWaitTime * time.Second)

		// tests
		t.Run("should create k8s service", func(t *testing.T) {
			k8sService, err := k8sResourcesClient.GetService(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sService(t, k8sService, resourceName, expectedLabels, v1core.ProtocolTCP, 80, 8080)
		})

		t.Run("should create k8s secret with client credentials", func(t *testing.T) {
			k8sSecret, err := k8sResourcesClient.GetSecret(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sBasicAuthSecret(t, k8sSecret, resourceName, expectedLabels, "username", "password")
		})

		t.Run("should create k8s secret with request parameters", func(t *testing.T) {
			k8sSecret, err := k8sResourcesClient.GetSecret(paramsSecretName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sParamsSecret(t, k8sSecret, paramsSecretName, expectedLabels, "headerKey", "headerValue", "queryParameterKey", "queryParameterValue")
		})

		t.Run("should create istio denier", func(t *testing.T) {
			denier, err := k8sResourcesClient.GetHandler(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sIstioHandler(t, denier, resourceName, expectedLabels, 7, "Not allowed")
		})

		t.Run("should create istio rule", func(t *testing.T) {
			rule, err := k8sResourcesClient.GetRule(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sIstioRule(t, rule, resourceName, config.Namespace, expectedLabels)
		})

		t.Run("should create istio checknothing", func(t *testing.T) {
			checknothing, err := k8sResourcesClient.GetInstance(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sIstioInstance(t, checknothing, resourceName, expectedLabels)
		})

		t.Run("should add service to application custom resource", func(t *testing.T) {
			application, err := k8sResourcesClient.GetApplicationServices(dummyApp.Name, v1.GetOptions{})
			require.NoError(t, err)

			expectedCSRF := "not used in this test run"
			if api.Credentials.Basic.CSRFInfo != nil {
				expectedCSRF = api.Credentials.Basic.CSRFInfo.TokenEndpointURL
			}

			expectedServiceData := testkit.ServiceData{
				ServiceId:            serviceId,
				DisplayName:          "test service",
				ProviderDisplayName:  "service provider",
				LongDescription:      "service description",
				HasAPI:               true,
				TargetUrl:            "http://service.com",
				OauthUrl:             "",
				GatewayUrl:           "http://" + resourceName + ".kyma-integration.svc.cluster.local",
				AccessLabel:          resourceName,
				HasEvents:            false,
				CSRFTokenEndpointURL: expectedCSRF,
			}

			testkit.CheckK8sApplication(t, application, dummyApp.Name, expectedServiceData)
		})
	}

	t.Run("when creating service only with Basic Auth API", func(t *testing.T) {
		t.Run("when current api with request parameters provided", func(t *testing.T) {
			testWithBasicAuth(defaultBasicAPI.WithRequestParameters(&requestParameters), t)
		})

		t.Run("when deprecated api provided", func(t *testing.T) {
			testWithBasicAuth(defaultBasicAPI.WithHeadersAndQueryParameters(&headers, &queryParameters), t)
		})
	})

	t.Run("when creating service only with Basic Auth API and CSRF", func(t *testing.T) {
		csrfInfo := &testkit.CSRFInfo{TokenEndpointURL: "https://csrf.token.endpoint.org"}

		t.Run("when current api with request parameters provided", func(t *testing.T) {
			testWithBasicAuth(defaultBasicAPI.WithRequestParameters(&requestParameters).WithCSRFInBasic(csrfInfo), t)
		})

		t.Run("when deprecated api provided", func(t *testing.T) {
			testWithBasicAuth(defaultBasicAPI.WithHeadersAndQueryParameters(&headers, &queryParameters).WithCSRFInBasic(csrfInfo), t)
		})
	})

	defaultCertificateGenAPI := testkit.API{
		TargetUrl: "http://service.com",
		Credentials: &testkit.Credentials{
			CertificateGen: &testkit.CertificateGen{
				CommonName: "commonName",
			},
		},
		Spec: testkit.ApiRawSpec,
	}

	testWithCertGen := func(api *testkit.API, t *testing.T) {

		// setup
		initialServiceDefinition := testkit.ServiceDetails{
			Name:        "test service",
			Provider:    "service provider",
			Description: "service description",
			Api:         api,
		}

		statusCode, postResponseData, err := metadataServiceClient.CreateService(t, initialServiceDefinition)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		serviceId := postResponseData.ID
		defer metadataServiceClient.DeleteService(t, serviceId)

		resourceName := dummyApp.Name + "-" + serviceId
		paramsSecretName := testkit.CreateParamsSecretName(resourceName)

		expectedLabels := map[string]string{"app": dummyApp.Name, "serviceId": serviceId}

		time.Sleep(crPropagationWaitTime * time.Second)

		// tests
		t.Run("should create k8s service", func(t *testing.T) {
			k8sService, err := k8sResourcesClient.GetService(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sService(t, k8sService, resourceName, expectedLabels, v1core.ProtocolTCP, 80, 8080)
		})

		t.Run("should create k8s secret with client credentials", func(t *testing.T) {
			k8sSecret, err := k8sResourcesClient.GetSecret(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sCertificateGenSecret(t, k8sSecret, resourceName, expectedLabels, "commonName")
		})

		t.Run("should create k8s secret with request parameters", func(t *testing.T) {
			k8sSecret, err := k8sResourcesClient.GetSecret(paramsSecretName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sParamsSecret(t, k8sSecret, paramsSecretName, expectedLabels, "headerKey", "headerValue", "queryParameterKey", "queryParameterValue")
		})

		t.Run("should create istio denier", func(t *testing.T) {
			denier, err := k8sResourcesClient.GetHandler(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sIstioHandler(t, denier, resourceName, expectedLabels, 7, "Not allowed")
		})

		t.Run("should create istio rule", func(t *testing.T) {
			rule, err := k8sResourcesClient.GetRule(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sIstioRule(t, rule, resourceName, config.Namespace, expectedLabels)
		})

		t.Run("should create istio checknothing", func(t *testing.T) {
			checknothing, err := k8sResourcesClient.GetInstance(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sIstioInstance(t, checknothing, resourceName, expectedLabels)
		})

		t.Run("should add service to application custom resource", func(t *testing.T) {
			application, err := k8sResourcesClient.GetApplicationServices(dummyApp.Name, v1.GetOptions{})
			require.NoError(t, err)

			expectedCSRF := "not used in this test run"
			if api.Credentials.CertificateGen.CSRFInfo != nil {
				expectedCSRF = api.Credentials.CertificateGen.CSRFInfo.TokenEndpointURL
			}

			expectedServiceData := testkit.ServiceData{
				ServiceId:            serviceId,
				DisplayName:          "test service",
				ProviderDisplayName:  "service provider",
				LongDescription:      "service description",
				HasAPI:               true,
				TargetUrl:            "http://service.com",
				OauthUrl:             "",
				GatewayUrl:           "http://" + resourceName + ".kyma-integration.svc.cluster.local",
				AccessLabel:          resourceName,
				HasEvents:            false,
				CSRFTokenEndpointURL: expectedCSRF,
			}

			testkit.CheckK8sApplication(t, application, dummyApp.Name, expectedServiceData)
		})
	}

	t.Run("when creating service only with CertificateGen API", func(t *testing.T) {
		t.Run("when current api with request parameters provided", func(t *testing.T) {
			testWithCertGen(defaultCertificateGenAPI.WithRequestParameters(&requestParameters), t)
		})

		t.Run("when deprecated api provided", func(t *testing.T) {
			testWithCertGen(defaultCertificateGenAPI.WithHeadersAndQueryParameters(&headers, &queryParameters), t)
		})
	})

	t.Run("when creating service only with CertificateGen API and CSRF", func(t *testing.T) {
		csrfInfo := &testkit.CSRFInfo{TokenEndpointURL: "https://csrf.token.endpoint.org"}

		t.Run("when current api with request parameters provided", func(t *testing.T) {
			testWithCertGen(defaultCertificateGenAPI.WithRequestParameters(&requestParameters).WithCSRFInCertificateGen(csrfInfo), t)
		})

		t.Run("when deprecated api provided", func(t *testing.T) {
			testWithCertGen(defaultCertificateGenAPI.WithHeadersAndQueryParameters(&headers, &queryParameters).WithCSRFInCertificateGen(csrfInfo), t)
		})
	})

	t.Run("when creating service only with Events", func(t *testing.T) {

		// setup
		initialServiceDefinition := testkit.ServiceDetails{
			Name:        "test service",
			Provider:    "service provider",
			Description: "service description",
			Events: &testkit.Events{
				Spec: testkit.EventsRawSpec,
			},
		}

		statusCode, postResponseData, err := metadataServiceClient.CreateService(t, initialServiceDefinition)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		serviceId := postResponseData.ID
		defer metadataServiceClient.DeleteService(t, serviceId)

		resourceName := dummyApp.Name + "-" + serviceId

		time.Sleep(crPropagationWaitTime * time.Second)

		// tests
		t.Run("should not create k8s service", func(t *testing.T) {
			_, err := k8sResourcesClient.GetService(resourceName, v1.GetOptions{})
			require.Error(t, err)
			require.True(t, k8serrors.IsNotFound(err))
		})

		t.Run("should not create k8s secret with client credentials", func(t *testing.T) {
			_, err := k8sResourcesClient.GetSecret(resourceName, v1.GetOptions{})
			require.Error(t, err)
			require.True(t, k8serrors.IsNotFound(err))
		})

		t.Run("should not create istio denier", func(t *testing.T) {
			_, err := k8sResourcesClient.GetHandler(resourceName, v1.GetOptions{})
			require.Error(t, err)
			require.True(t, k8serrors.IsNotFound(err))
		})

		t.Run("should not create istio rule", func(t *testing.T) {
			_, err := k8sResourcesClient.GetRule(resourceName, v1.GetOptions{})
			require.Error(t, err)
			require.True(t, k8serrors.IsNotFound(err))
		})

		t.Run("should not create istio checknothing", func(t *testing.T) {
			_, err := k8sResourcesClient.GetInstance(resourceName, v1.GetOptions{})
			require.Error(t, err)
			require.True(t, k8serrors.IsNotFound(err))
		})

		t.Run("should add service to application custom resource", func(t *testing.T) {
			application, err := k8sResourcesClient.GetApplicationServices(dummyApp.Name, v1.GetOptions{})
			require.NoError(t, err)

			expectedServiceData := testkit.ServiceData{
				ServiceId:           serviceId,
				DisplayName:         "test service",
				ProviderDisplayName: "service provider",
				LongDescription:     "service description",
				HasAPI:              false,
				HasEvents:           true,
			}

			testkit.CheckK8sApplication(t, application, dummyApp.Name, expectedServiceData)
		})
	})

	t.Run("when updating service and changing API plus adding Events", func(t *testing.T) {

		// setup
		initialServiceDefinition := testkit.ServiceDetails{
			Name:        "test service",
			Provider:    "service provider",
			Description: "service description",
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
		}

		statusCode, postResponseData, err := metadataServiceClient.CreateService(t, initialServiceDefinition)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		serviceId := postResponseData.ID
		defer metadataServiceClient.DeleteService(t, serviceId)

		resourceName := dummyApp.Name + "-" + serviceId

		expectedLabels := map[string]string{"app": dummyApp.Name, "serviceId": serviceId}

		updatedServiceDefinition := testkit.ServiceDetails{
			Name:        "updated test service",
			Provider:    "updated service provider",
			Description: "updated service description",
			Api: &testkit.API{
				TargetUrl: "http://updated-service.com",
				Credentials: &testkit.Credentials{
					Oauth: &testkit.Oauth{
						URL:          "http://updated-oauth.com",
						ClientID:     "updated-clientId",
						ClientSecret: "updated-clientSecret",
					},
				},
				Spec: testkit.ApiRawSpec,
			},
			Events: &testkit.Events{
				Spec: testkit.EventsRawSpec,
			},
		}

		statusCode, err = metadataServiceClient.UpdateService(t, serviceId, updatedServiceDefinition)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		time.Sleep(crPropagationWaitTime * time.Second)

		// tests
		t.Run("should preserve k8s service", func(t *testing.T) {
			k8sService, err := k8sResourcesClient.GetService(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sService(t, k8sService, resourceName, expectedLabels, v1core.ProtocolTCP, 80, 8080)
		})

		t.Run("should update k8s secret with client credentials", func(t *testing.T) {
			k8sSecret, err := k8sResourcesClient.GetSecret(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sOAuthSecret(t, k8sSecret, resourceName, expectedLabels, "updated-clientId", "updated-clientSecret")
		})

		t.Run("should preserve istio denier", func(t *testing.T) {
			denier, err := k8sResourcesClient.GetHandler(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sIstioHandler(t, denier, resourceName, expectedLabels, 7, "Not allowed")
		})

		t.Run("should preserve istio rule", func(t *testing.T) {
			rule, err := k8sResourcesClient.GetRule(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sIstioRule(t, rule, resourceName, config.Namespace, expectedLabels)
		})

		t.Run("should preserve istio checknothing", func(t *testing.T) {
			checknothing, err := k8sResourcesClient.GetInstance(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sIstioInstance(t, checknothing, resourceName, expectedLabels)
		})

		t.Run("should update service inside application custom resource", func(t *testing.T) {
			application, err := k8sResourcesClient.GetApplicationServices(dummyApp.Name, v1.GetOptions{})
			require.NoError(t, err)

			expectedServiceData := testkit.ServiceData{
				ServiceId:           serviceId,
				DisplayName:         "updated test service",
				ProviderDisplayName: "updated service provider",
				LongDescription:     "updated service description",
				HasAPI:              true,
				TargetUrl:           "http://updated-service.com",
				OauthUrl:            "http://updated-oauth.com",
				GatewayUrl:          "http://" + resourceName + ".kyma-integration.svc.cluster.local",
				AccessLabel:         resourceName,
				HasEvents:           true,
			}

			testkit.CheckK8sApplication(t, application, dummyApp.Name, expectedServiceData)
		})
	})

	t.Run("when updating service and removing API", func(t *testing.T) {

		// setup
		initialServiceDefinition := testkit.ServiceDetails{
			Name:        "test service",
			Provider:    "service provider",
			Description: "service description",
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
		}

		statusCode, postResponseData, err := metadataServiceClient.CreateService(t, initialServiceDefinition)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		serviceId := postResponseData.ID
		defer metadataServiceClient.DeleteService(t, serviceId)

		resourceName := dummyApp.Name + "-" + serviceId

		updatedServiceDefinition := testkit.ServiceDetails{
			Name:        "updated test service",
			Provider:    "updated service provider",
			Description: "updated service description",
			Events: &testkit.Events{
				Spec: testkit.EventsRawSpec,
			},
		}

		statusCode, err = metadataServiceClient.UpdateService(t, serviceId, updatedServiceDefinition)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		time.Sleep(crPropagationWaitTime * time.Second)

		// tests
		t.Run("should remove k8s service", func(t *testing.T) {
			_, err := k8sResourcesClient.GetService(resourceName, v1.GetOptions{})
			require.Error(t, err)
			require.True(t, k8serrors.IsNotFound(err))
		})

		t.Run("should remove k8s secret with client credentials", func(t *testing.T) {
			_, err := k8sResourcesClient.GetSecret(resourceName, v1.GetOptions{})
			require.Error(t, err)
			require.True(t, k8serrors.IsNotFound(err))
		})

		t.Run("should remove istio denier", func(t *testing.T) {
			_, err := k8sResourcesClient.GetHandler(resourceName, v1.GetOptions{})
			require.Error(t, err)
			require.True(t, k8serrors.IsNotFound(err))
		})

		t.Run("should remove istio rule", func(t *testing.T) {
			_, err := k8sResourcesClient.GetRule(resourceName, v1.GetOptions{})
			require.Error(t, err)
			require.True(t, k8serrors.IsNotFound(err))
		})

		t.Run("should remove istio checknothing", func(t *testing.T) {
			_, err := k8sResourcesClient.GetInstance(resourceName, v1.GetOptions{})
			require.Error(t, err)
			require.True(t, k8serrors.IsNotFound(err))
		})

		t.Run("should update service inside application custom resource", func(t *testing.T) {
			application, err := k8sResourcesClient.GetApplicationServices(dummyApp.Name, v1.GetOptions{})
			require.NoError(t, err)

			expectedServiceData := testkit.ServiceData{
				ServiceId:           serviceId,
				DisplayName:         "updated test service",
				ProviderDisplayName: "updated service provider",
				LongDescription:     "updated service description",
				HasAPI:              false,
				HasEvents:           true,
			}

			testkit.CheckK8sApplication(t, application, dummyApp.Name, expectedServiceData)
		})
	})

	t.Run("when updating service and adding OAuth API", func(t *testing.T) {

		// setup
		initialServiceDefinition := testkit.ServiceDetails{
			Name:        "test service",
			Provider:    "service provider",
			Description: "service description",
			Events: &testkit.Events{
				Spec: testkit.EventsRawSpec,
			},
		}

		statusCode, postResponseData, err := metadataServiceClient.CreateService(t, initialServiceDefinition)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		serviceId := postResponseData.ID
		defer metadataServiceClient.DeleteService(t, serviceId)

		resourceName := dummyApp.Name + "-" + serviceId

		expectedLabels := map[string]string{"app": dummyApp.Name, "serviceId": serviceId}

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
		}

		statusCode, err = metadataServiceClient.UpdateService(t, serviceId, updatedServiceDefinition)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		time.Sleep(crPropagationWaitTime * time.Second)

		// tests
		t.Run("should create k8s service", func(t *testing.T) {
			k8sService, err := k8sResourcesClient.GetService(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sService(t, k8sService, resourceName, expectedLabels, v1core.ProtocolTCP, 80, 8080)
		})

		t.Run("should create k8s secret with client credentials", func(t *testing.T) {
			k8sSecret, err := k8sResourcesClient.GetSecret(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sOAuthSecret(t, k8sSecret, resourceName, expectedLabels, "clientId", "clientSecret")
		})

		t.Run("should create istio denier", func(t *testing.T) {
			denier, err := k8sResourcesClient.GetHandler(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sIstioHandler(t, denier, resourceName, expectedLabels, 7, "Not allowed")
		})

		t.Run("should create istio rule", func(t *testing.T) {
			rule, err := k8sResourcesClient.GetRule(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sIstioRule(t, rule, resourceName, config.Namespace, expectedLabels)
		})

		t.Run("should create istio checknothing", func(t *testing.T) {
			checknothing, err := k8sResourcesClient.GetInstance(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sIstioInstance(t, checknothing, resourceName, expectedLabels)
		})

		t.Run("should update service inside application custom resource", func(t *testing.T) {
			application, err := k8sResourcesClient.GetApplicationServices(dummyApp.Name, v1.GetOptions{})
			require.NoError(t, err)

			expectedServiceData := testkit.ServiceData{
				ServiceId:           serviceId,
				DisplayName:         "updated test service",
				ProviderDisplayName: "updated service provider",
				LongDescription:     "updated service description",
				HasAPI:              true,
				TargetUrl:           "http://service.com",
				OauthUrl:            "http://oauth.com",
				GatewayUrl:          "http://" + resourceName + ".kyma-integration.svc.cluster.local",
				AccessLabel:         resourceName,
				HasEvents:           false,
			}

			testkit.CheckK8sApplication(t, application, dummyApp.Name, expectedServiceData)
		})
	})

	t.Run("when updating service and adding Basic Auth API", func(t *testing.T) {

		// setup
		initialServiceDefinition := testkit.ServiceDetails{
			Name:        "test service",
			Provider:    "service provider",
			Description: "service description",
			Events: &testkit.Events{
				Spec: testkit.EventsRawSpec,
			},
		}

		statusCode, postResponseData, err := metadataServiceClient.CreateService(t, initialServiceDefinition)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		serviceId := postResponseData.ID
		defer metadataServiceClient.DeleteService(t, serviceId)

		resourceName := dummyApp.Name + "-" + serviceId

		expectedLabels := map[string]string{"app": dummyApp.Name, "serviceId": serviceId}

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
		}

		statusCode, err = metadataServiceClient.UpdateService(t, serviceId, updatedServiceDefinition)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		time.Sleep(crPropagationWaitTime * time.Second)

		// tests
		t.Run("should create k8s service", func(t *testing.T) {
			k8sService, err := k8sResourcesClient.GetService(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sService(t, k8sService, resourceName, expectedLabels, v1core.ProtocolTCP, 80, 8080)
		})

		t.Run("should create k8s secret with client credentials", func(t *testing.T) {
			k8sSecret, err := k8sResourcesClient.GetSecret(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sBasicAuthSecret(t, k8sSecret, resourceName, expectedLabels, "username", "password")
		})

		t.Run("should create istio denier", func(t *testing.T) {
			denier, err := k8sResourcesClient.GetHandler(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sIstioHandler(t, denier, resourceName, expectedLabels, 7, "Not allowed")
		})

		t.Run("should create istio rule", func(t *testing.T) {
			rule, err := k8sResourcesClient.GetRule(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sIstioRule(t, rule, resourceName, config.Namespace, expectedLabels)
		})

		t.Run("should create istio checknothing", func(t *testing.T) {
			checknothing, err := k8sResourcesClient.GetInstance(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sIstioInstance(t, checknothing, resourceName, expectedLabels)
		})

		t.Run("should update service inside application custom resource", func(t *testing.T) {
			application, err := k8sResourcesClient.GetApplicationServices(dummyApp.Name, v1.GetOptions{})
			require.NoError(t, err)

			expectedServiceData := testkit.ServiceData{
				ServiceId:           serviceId,
				DisplayName:         "updated test service",
				ProviderDisplayName: "updated service provider",
				LongDescription:     "updated service description",
				HasAPI:              true,
				TargetUrl:           "http://service.com",
				OauthUrl:            "",
				GatewayUrl:          "http://" + resourceName + ".kyma-integration.svc.cluster.local",
				AccessLabel:         resourceName,
				HasEvents:           false,
			}

			testkit.CheckK8sApplication(t, application, dummyApp.Name, expectedServiceData)
		})
	})

	t.Run("when updating service and adding CertificateGen API", func(t *testing.T) {

		// setup
		initialServiceDefinition := testkit.ServiceDetails{
			Name:        "test service",
			Provider:    "service provider",
			Description: "service description",
			Events: &testkit.Events{
				Spec: testkit.EventsRawSpec,
			},
		}

		statusCode, postResponseData, err := metadataServiceClient.CreateService(t, initialServiceDefinition)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		serviceId := postResponseData.ID
		defer metadataServiceClient.DeleteService(t, serviceId)

		resourceName := dummyApp.Name + "-" + serviceId

		expectedLabels := map[string]string{"app": dummyApp.Name, "serviceId": serviceId}

		updatedServiceDefinition := testkit.ServiceDetails{
			Name:        "updated test service",
			Provider:    "updated service provider",
			Description: "updated service description",
			Api: &testkit.API{
				TargetUrl: "http://service.com",
				Credentials: &testkit.Credentials{
					CertificateGen: &testkit.CertificateGen{
						CommonName: "commonName",
					},
				},
				Spec: testkit.ApiRawSpec,
			},
		}

		statusCode, err = metadataServiceClient.UpdateService(t, serviceId, updatedServiceDefinition)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		time.Sleep(crPropagationWaitTime * time.Second)

		// tests
		t.Run("should create k8s service", func(t *testing.T) {
			k8sService, err := k8sResourcesClient.GetService(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sService(t, k8sService, resourceName, expectedLabels, v1core.ProtocolTCP, 80, 8080)
		})

		t.Run("should create k8s secret with client credentials", func(t *testing.T) {
			k8sSecret, err := k8sResourcesClient.GetSecret(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sCertificateGenSecret(t, k8sSecret, resourceName, expectedLabels, "commonName")
		})

		t.Run("should create istio denier", func(t *testing.T) {
			denier, err := k8sResourcesClient.GetHandler(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sIstioHandler(t, denier, resourceName, expectedLabels, 7, "Not allowed")
		})

		t.Run("should create istio rule", func(t *testing.T) {
			rule, err := k8sResourcesClient.GetRule(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sIstioRule(t, rule, resourceName, config.Namespace, expectedLabels)
		})

		t.Run("should create istio checknothing", func(t *testing.T) {
			checknothing, err := k8sResourcesClient.GetInstance(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sIstioInstance(t, checknothing, resourceName, expectedLabels)
		})

		t.Run("should update service inside application custom resource", func(t *testing.T) {
			application, err := k8sResourcesClient.GetApplicationServices(dummyApp.Name, v1.GetOptions{})
			require.NoError(t, err)

			expectedServiceData := testkit.ServiceData{
				ServiceId:           serviceId,
				DisplayName:         "updated test service",
				ProviderDisplayName: "updated service provider",
				LongDescription:     "updated service description",
				HasAPI:              true,
				TargetUrl:           "http://service.com",
				OauthUrl:            "",
				GatewayUrl:          "http://" + resourceName + ".kyma-integration.svc.cluster.local",
				AccessLabel:         resourceName,
				HasEvents:           false,
			}

			testkit.CheckK8sApplication(t, application, dummyApp.Name, expectedServiceData)
		})
	})

	t.Run("when updating service with CertificateGen API without changing CN", func(t *testing.T) {

		// setup
		initialServiceDefinition := testkit.ServiceDetails{
			Name:        "test service",
			Provider:    "service provider",
			Description: "service description",
			Api: &testkit.API{
				TargetUrl: "http://service.com",
				Credentials: &testkit.Credentials{
					CertificateGen: &testkit.CertificateGen{
						CommonName: "commonName",
					},
				},
				Spec: testkit.ApiRawSpec,
			},
		}

		statusCode, postResponseData, err := metadataServiceClient.CreateService(t, initialServiceDefinition)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		serviceId := postResponseData.ID
		defer metadataServiceClient.DeleteService(t, serviceId)

		resourceName := dummyApp.Name + "-" + serviceId

		expectedLabels := map[string]string{"app": dummyApp.Name, "serviceId": serviceId}

		var initialSecret *v1core.Secret
		var error error

		t.Run("should create k8s initialSecret with certificate", func(t *testing.T) {
			initialSecret, error = k8sResourcesClient.GetSecret(resourceName, v1.GetOptions{})
			require.NoError(t, error)

			testkit.CheckK8sCertificateGenSecret(t, initialSecret, resourceName, expectedLabels, "commonName")
		})

		updatedServiceDefinition := testkit.ServiceDetails{
			Name:        "updated test service",
			Provider:    "updated service provider",
			Description: "updated service description",
			Api: &testkit.API{
				TargetUrl: "http://service.com",
				Credentials: &testkit.Credentials{
					CertificateGen: &testkit.CertificateGen{
						CommonName: "commonName",
					},
				},
				Spec: testkit.ApiRawSpec,
			},
		}

		statusCode, err = metadataServiceClient.UpdateService(t, serviceId, updatedServiceDefinition)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		time.Sleep(crPropagationWaitTime * time.Second)

		t.Run("should not override certificate in k8s initialSecret", func(t *testing.T) {
			updatedSecret, err := k8sResourcesClient.GetSecret(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sCertificateGenSecret(t, updatedSecret, resourceName, expectedLabels, "commonName")

			require.Equal(t, initialSecret, updatedSecret)
		})

		// tests
		t.Run("should create k8s service", func(t *testing.T) {
			k8sService, err := k8sResourcesClient.GetService(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sService(t, k8sService, resourceName, expectedLabels, v1core.ProtocolTCP, 80, 8080)
		})

		t.Run("should create istio denier", func(t *testing.T) {
			denier, err := k8sResourcesClient.GetHandler(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sIstioHandler(t, denier, resourceName, expectedLabels, 7, "Not allowed")
		})

		t.Run("should create istio rule", func(t *testing.T) {
			rule, err := k8sResourcesClient.GetRule(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sIstioRule(t, rule, resourceName, config.Namespace, expectedLabels)
		})

		t.Run("should create istio checknothing", func(t *testing.T) {
			checknothing, err := k8sResourcesClient.GetInstance(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sIstioInstance(t, checknothing, resourceName, expectedLabels)
		})

		t.Run("should update service inside application custom resource", func(t *testing.T) {
			application, err := k8sResourcesClient.GetApplicationServices(dummyApp.Name, v1.GetOptions{})
			require.NoError(t, err)

			expectedServiceData := testkit.ServiceData{
				ServiceId:           serviceId,
				DisplayName:         "updated test service",
				ProviderDisplayName: "updated service provider",
				LongDescription:     "updated service description",
				HasAPI:              true,
				TargetUrl:           "http://service.com",
				OauthUrl:            "",
				GatewayUrl:          "http://" + resourceName + ".kyma-integration.svc.cluster.local",
				AccessLabel:         resourceName,
				HasEvents:           false,
			}

			testkit.CheckK8sApplication(t, application, dummyApp.Name, expectedServiceData)
		})
	})

	t.Run("when deleting service", func(t *testing.T) {

		// setup
		initialServiceDefinition := testkit.ServiceDetails{
			Name:        "test service",
			Provider:    "service provider",
			Description: "service description",
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
		}

		statusCode, postResponseData, err := metadataServiceClient.CreateService(t, initialServiceDefinition)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		serviceId := postResponseData.ID
		resourceName := dummyApp.Name + "-" + serviceId

		statusCode, err = metadataServiceClient.DeleteService(t, serviceId)
		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, statusCode)

		time.Sleep(crPropagationWaitTime * time.Second)

		// tests
		t.Run("should remove k8s service", func(t *testing.T) {
			_, err := k8sResourcesClient.GetService(resourceName, v1.GetOptions{})
			require.Error(t, err)
			require.True(t, k8serrors.IsNotFound(err))
		})

		t.Run("should remove k8s secret with client credentials", func(t *testing.T) {
			_, err := k8sResourcesClient.GetSecret(resourceName, v1.GetOptions{})
			require.Error(t, err)
			require.True(t, k8serrors.IsNotFound(err))
		})

		t.Run("should remove istio denier", func(t *testing.T) {
			_, err := k8sResourcesClient.GetHandler(resourceName, v1.GetOptions{})
			require.Error(t, err)
			require.True(t, k8serrors.IsNotFound(err))
		})

		t.Run("should remove istio rule", func(t *testing.T) {
			_, err := k8sResourcesClient.GetRule(resourceName, v1.GetOptions{})
			require.Error(t, err)
			require.True(t, k8serrors.IsNotFound(err))
		})

		t.Run("should remove istio checknothing", func(t *testing.T) {
			_, err := k8sResourcesClient.GetInstance(resourceName, v1.GetOptions{})
			require.Error(t, err)
			require.True(t, k8serrors.IsNotFound(err))
		})

		t.Run("should remove service from application custom resource", func(t *testing.T) {
			application, err := k8sResourcesClient.GetApplicationServices(dummyApp.Name, v1.GetOptions{})
			require.NoError(t, err)
			testkit.CheckK8sApplicationNotContainsService(t, application, serviceId)
		})
	})

	err = k8sResourcesClient.DeleteApplication(dummyApp.Name, &v1.DeleteOptions{})
	require.NoError(t, err)
}

func TestK8sApplicationDeletion(t *testing.T) {

	config, err := testkit.ReadConfig()
	require.NoError(t, err)

	k8sResourcesClient, err := testkit.NewK8sInClusterResourcesClient(config.Namespace)
	require.NoError(t, err)

	dummyApp, err := k8sResourcesClient.CreateDummyApplication("appk8srestest1", v1.GetOptions{}, true)
	require.NoError(t, err)

	defer k8sResourcesClient.DeleteApplication(dummyApp.Name, &v1.DeleteOptions{})

	metadataServiceClient := testkit.NewMetadataServiceClient(config.MetadataServiceUrl + "/" + dummyApp.Name + "/v1/metadata/services")

	t.Run("when deleting application", func(t *testing.T) {
		//given
		initialServiceDefinition := testkit.ServiceDetails{
			Name:        "test service",
			Provider:    "service provider",
			Description: "service description",
			Api: &testkit.API{
				TargetUrl: "http://service.com",
				Credentials: &testkit.Credentials{
					Oauth: &testkit.Oauth{
						URL:          "http://oauth.com",
						ClientID:     "clientId",
						ClientSecret: "clientSecret",
					},
				},
				RequestParameters: &requestParameters,
				Spec:              testkit.ApiRawSpec,
			},
			Events: &testkit.Events{
				Spec: testkit.EventsRawSpec,
			},
		}

		statusCode, postResponseData, err := metadataServiceClient.CreateService(t, initialServiceDefinition)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		serviceId := postResponseData.ID
		defer metadataServiceClient.DeleteService(t, serviceId)

		resourceName := dummyApp.Name + "-" + serviceId
		paramsSecretName := testkit.CreateParamsSecretName(resourceName)

		//when
		err = k8sResourcesClient.DeleteApplication(dummyApp.Name, &v1.DeleteOptions{})
		require.NoError(t, err)

		time.Sleep(deleteApplicationResourcesWaitTime * time.Second)

		//then
		t.Run("should remove k8s service", func(t *testing.T) {
			_, err := k8sResourcesClient.GetService(resourceName, v1.GetOptions{})
			require.Error(t, err)
			require.True(t, k8serrors.IsNotFound(err))
		})

		t.Run("should remove k8s secret with client credentials", func(t *testing.T) {
			_, err := k8sResourcesClient.GetSecret(resourceName, v1.GetOptions{})
			require.Error(t, err)
			require.True(t, k8serrors.IsNotFound(err))
		})

		t.Run("should remove k8s secret with request parameters", func(t *testing.T) {
			_, err := k8sResourcesClient.GetSecret(paramsSecretName, v1.GetOptions{})
			require.Error(t, err)
			require.True(t, k8serrors.IsNotFound(err))
		})

		t.Run("should remove istio denier", func(t *testing.T) {
			_, err := k8sResourcesClient.GetHandler(resourceName, v1.GetOptions{})
			require.Error(t, err)
			require.True(t, k8serrors.IsNotFound(err))
		})

		t.Run("should remove istio rule", func(t *testing.T) {
			_, err := k8sResourcesClient.GetRule(resourceName, v1.GetOptions{})
			require.Error(t, err)
			require.True(t, k8serrors.IsNotFound(err))
		})

		t.Run("should remove istio checknothing", func(t *testing.T) {
			_, err := k8sResourcesClient.GetInstance(resourceName, v1.GetOptions{})
			require.Error(t, err)
			require.True(t, k8serrors.IsNotFound(err))
		})
	})
}
