package k8stests

import (
	"github.com/stretchr/testify/require"
	v1core "k8s.io/api/core/v1"
	"time"

	"net/http"
	"testing"

	"github.com/kyma-project/kyma/tests/metadata-service-tests/test/testkit"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

const crPropagationWaitTime = 10

func TestK8sResources(t *testing.T) {

	config, err := testkit.ReadConfig()
	require.NoError(t, err)

	k8sResourcesClient, err := testkit.NewK8sInClusterResourcesClient(config.Namespace)
	require.NoError(t, err)

	dummyRE, err := k8sResourcesClient.CreateDummyRemoteEnvironment("dummy-re", v1.GetOptions{})
	require.NoError(t, err)

	metadataServiceClient := testkit.NewMetadataServiceClient(config.MetadataServiceUrl + "/" + dummyRE.Name + "/v1/metadata/services")

	t.Run("when creating service only with OAuth API", func(t *testing.T) {

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
		resourceName := "re-" + dummyRE.Name + "-" + serviceId

		expectedLabels := map[string]string{"re": dummyRE.Name, "serviceId": serviceId}

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
			denier, err := k8sResourcesClient.GetDenier(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sIstioDenier(t, denier, resourceName, expectedLabels, 7, "Not allowed")
		})

		t.Run("should create istio rule", func(t *testing.T) {
			rule, err := k8sResourcesClient.GetRule(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sIstioRule(t, rule, resourceName, config.Namespace, expectedLabels)
		})

		t.Run("should create istio checknothing", func(t *testing.T) {
			checknothing, err := k8sResourcesClient.GetChecknothing(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sChecknothing(t, checknothing, resourceName, expectedLabels)
		})

		t.Run("should add service to remote environment custom resource", func(t *testing.T) {
			remoteEnvironment, err := k8sResourcesClient.GetRemoteEnvironmentServices(dummyRE.Name, v1.GetOptions{})
			require.NoError(t, err)

			expectedServiceData := testkit.ServiceData{
				ServiceId:           serviceId,
				DisplayName:         "test service",
				ProviderDisplayName: "service provider",
				LongDescription:     "service description",
				HasAPI:              true,
				TargetUrl:           "http://service.com",
				OauthUrl:            "http://oauth.com",
				GatewayUrl:          "http://" + resourceName + ".kyma-integration.svc.cluster.local",
				AccessLabel:         resourceName,
				HasEvents:           false,
			}

			testkit.CheckK8sRemoteEnvironment(t, remoteEnvironment, dummyRE.Name, expectedServiceData)
		})

		// clean up
		metadataServiceClient.DeleteService(t, serviceId)
	})

	t.Run("when creating service only with Basic Auth API", func(t *testing.T) {

		// setup
		initialServiceDefinition := testkit.ServiceDetails{
			Name:        "test service",
			Provider:    "service provider",
			Description: "service description",
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

		statusCode, postResponseData, err := metadataServiceClient.CreateService(t, initialServiceDefinition)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, statusCode)

		serviceId := postResponseData.ID
		resourceName := "re-" + dummyRE.Name + "-" + serviceId

		expectedLabels := map[string]string{"re": dummyRE.Name, "serviceId": serviceId}

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
			denier, err := k8sResourcesClient.GetDenier(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sIstioDenier(t, denier, resourceName, expectedLabels, 7, "Not allowed")
		})

		t.Run("should create istio rule", func(t *testing.T) {
			rule, err := k8sResourcesClient.GetRule(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sIstioRule(t, rule, resourceName, config.Namespace, expectedLabels)
		})

		t.Run("should create istio checknothing", func(t *testing.T) {
			checknothing, err := k8sResourcesClient.GetChecknothing(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sChecknothing(t, checknothing, resourceName, expectedLabels)
		})

		t.Run("should add service to remote environment custom resource", func(t *testing.T) {
			remoteEnvironment, err := k8sResourcesClient.GetRemoteEnvironmentServices(dummyRE.Name, v1.GetOptions{})
			require.NoError(t, err)

			expectedServiceData := testkit.ServiceData{
				ServiceId:           serviceId,
				DisplayName:         "test service",
				ProviderDisplayName: "service provider",
				LongDescription:     "service description",
				HasAPI:              true,
				TargetUrl:           "http://service.com",
				OauthUrl:            "",
				GatewayUrl:          "http://" + resourceName + ".kyma-integration.svc.cluster.local",
				AccessLabel:         resourceName,
				HasEvents:           false,
			}

			testkit.CheckK8sRemoteEnvironment(t, remoteEnvironment, dummyRE.Name, expectedServiceData)
		})

		// clean up
		metadataServiceClient.DeleteService(t, serviceId)
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
		resourceName := "re-" + dummyRE.Name + "-" + serviceId

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
			_, err := k8sResourcesClient.GetDenier(resourceName, v1.GetOptions{})
			require.Error(t, err)
			require.True(t, k8serrors.IsNotFound(err))
		})

		t.Run("should not create istio rule", func(t *testing.T) {
			_, err := k8sResourcesClient.GetRule(resourceName, v1.GetOptions{})
			require.Error(t, err)
			require.True(t, k8serrors.IsNotFound(err))
		})

		t.Run("should not create istio checknothing", func(t *testing.T) {
			_, err := k8sResourcesClient.GetChecknothing(resourceName, v1.GetOptions{})
			require.Error(t, err)
			require.True(t, k8serrors.IsNotFound(err))
		})

		t.Run("should add service to remote environment custom resource", func(t *testing.T) {
			remoteEnvironment, err := k8sResourcesClient.GetRemoteEnvironmentServices(dummyRE.Name, v1.GetOptions{})
			require.NoError(t, err)

			expectedServiceData := testkit.ServiceData{
				ServiceId:           serviceId,
				DisplayName:         "test service",
				ProviderDisplayName: "service provider",
				LongDescription:     "service description",
				HasAPI:              false,
				HasEvents:           true,
			}

			testkit.CheckK8sRemoteEnvironment(t, remoteEnvironment, dummyRE.Name, expectedServiceData)
		})

		// clean up
		metadataServiceClient.DeleteService(t, serviceId)
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
		resourceName := "re-" + dummyRE.Name + "-" + serviceId

		expectedLabels := map[string]string{"re": dummyRE.Name, "serviceId": serviceId}

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
			denier, err := k8sResourcesClient.GetDenier(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sIstioDenier(t, denier, resourceName, expectedLabels, 7, "Not allowed")
		})

		t.Run("should preserve istio rule", func(t *testing.T) {
			rule, err := k8sResourcesClient.GetRule(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sIstioRule(t, rule, resourceName, config.Namespace, expectedLabels)
		})

		t.Run("should preserve istio checknothing", func(t *testing.T) {
			checknothing, err := k8sResourcesClient.GetChecknothing(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sChecknothing(t, checknothing, resourceName, expectedLabels)
		})

		t.Run("should update service inside remote environment custom resource", func(t *testing.T) {
			remoteEnvironment, err := k8sResourcesClient.GetRemoteEnvironmentServices(dummyRE.Name, v1.GetOptions{})
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

			testkit.CheckK8sRemoteEnvironment(t, remoteEnvironment, dummyRE.Name, expectedServiceData)
		})

		// clean up
		metadataServiceClient.DeleteService(t, serviceId)
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
		resourceName := "re-" + dummyRE.Name + "-" + serviceId

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
			_, err := k8sResourcesClient.GetDenier(resourceName, v1.GetOptions{})
			require.Error(t, err)
			require.True(t, k8serrors.IsNotFound(err))
		})

		t.Run("should remove istio rule", func(t *testing.T) {
			_, err := k8sResourcesClient.GetRule(resourceName, v1.GetOptions{})
			require.Error(t, err)
			require.True(t, k8serrors.IsNotFound(err))
		})

		t.Run("should remove istio checknothing", func(t *testing.T) {
			_, err := k8sResourcesClient.GetChecknothing(resourceName, v1.GetOptions{})
			require.Error(t, err)
			require.True(t, k8serrors.IsNotFound(err))
		})

		t.Run("should update service inside remote environment custom resource", func(t *testing.T) {
			remoteEnvironment, err := k8sResourcesClient.GetRemoteEnvironmentServices(dummyRE.Name, v1.GetOptions{})
			require.NoError(t, err)

			expectedServiceData := testkit.ServiceData{
				ServiceId:           serviceId,
				DisplayName:         "updated test service",
				ProviderDisplayName: "updated service provider",
				LongDescription:     "updated service description",
				HasAPI:              false,
				HasEvents:           true,
			}

			testkit.CheckK8sRemoteEnvironment(t, remoteEnvironment, dummyRE.Name, expectedServiceData)
		})

		// clean up
		metadataServiceClient.DeleteService(t, serviceId)
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
		resourceName := "re-" + dummyRE.Name + "-" + serviceId

		expectedLabels := map[string]string{"re": dummyRE.Name, "serviceId": serviceId}

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
			denier, err := k8sResourcesClient.GetDenier(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sIstioDenier(t, denier, resourceName, expectedLabels, 7, "Not allowed")
		})

		t.Run("should create istio rule", func(t *testing.T) {
			rule, err := k8sResourcesClient.GetRule(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sIstioRule(t, rule, resourceName, config.Namespace, expectedLabels)
		})

		t.Run("should create istio checknothing", func(t *testing.T) {
			checknothing, err := k8sResourcesClient.GetChecknothing(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sChecknothing(t, checknothing, resourceName, expectedLabels)
		})

		t.Run("should update service inside remote environment custom resource", func(t *testing.T) {
			remoteEnvironment, err := k8sResourcesClient.GetRemoteEnvironmentServices(dummyRE.Name, v1.GetOptions{})
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

			testkit.CheckK8sRemoteEnvironment(t, remoteEnvironment, dummyRE.Name, expectedServiceData)
		})

		// clean up
		metadataServiceClient.DeleteService(t, serviceId)
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
		resourceName := "re-" + dummyRE.Name + "-" + serviceId

		expectedLabels := map[string]string{"re": dummyRE.Name, "serviceId": serviceId}

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
			denier, err := k8sResourcesClient.GetDenier(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sIstioDenier(t, denier, resourceName, expectedLabels, 7, "Not allowed")
		})

		t.Run("should create istio rule", func(t *testing.T) {
			rule, err := k8sResourcesClient.GetRule(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sIstioRule(t, rule, resourceName, config.Namespace, expectedLabels)
		})

		t.Run("should create istio checknothing", func(t *testing.T) {
			checknothing, err := k8sResourcesClient.GetChecknothing(resourceName, v1.GetOptions{})
			require.NoError(t, err)

			testkit.CheckK8sChecknothing(t, checknothing, resourceName, expectedLabels)
		})

		t.Run("should update service inside remote environment custom resource", func(t *testing.T) {
			remoteEnvironment, err := k8sResourcesClient.GetRemoteEnvironmentServices(dummyRE.Name, v1.GetOptions{})
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

			testkit.CheckK8sRemoteEnvironment(t, remoteEnvironment, dummyRE.Name, expectedServiceData)
		})

		// clean up
		metadataServiceClient.DeleteService(t, serviceId)
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
		resourceName := "re-" + dummyRE.Name + "-" + serviceId

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
			_, err := k8sResourcesClient.GetDenier(resourceName, v1.GetOptions{})
			require.Error(t, err)
			require.True(t, k8serrors.IsNotFound(err))
		})

		t.Run("should remove istio rule", func(t *testing.T) {
			_, err := k8sResourcesClient.GetRule(resourceName, v1.GetOptions{})
			require.Error(t, err)
			require.True(t, k8serrors.IsNotFound(err))
		})

		t.Run("should remove istio checknothing", func(t *testing.T) {
			_, err := k8sResourcesClient.GetChecknothing(resourceName, v1.GetOptions{})
			require.Error(t, err)
			require.True(t, k8serrors.IsNotFound(err))
		})

		t.Run("should remove service from remote environment custom resource", func(t *testing.T) {
			remoteEnvironment, err := k8sResourcesClient.GetRemoteEnvironmentServices(dummyRE.Name, v1.GetOptions{})
			require.NoError(t, err)
			testkit.CheckK8sRemoteEnvironmentNotContainsService(t, remoteEnvironment, serviceId)
		})
	})

	err = k8sResourcesClient.DeleteRemoteEnvironment(dummyRE.Name, &v1.DeleteOptions{})
	require.NoError(t, err)
}
