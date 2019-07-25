package synchronization

import (
	"testing"

	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	k8smocks "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/k8sconsts/mocks"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestConverter(t *testing.T) {

	t.Run("should convert application without services", func(t *testing.T) {
		// given
		mockNameResolver := &k8smocks.NameResolver{}
		converter := NewConverter(mockNameResolver)

		description := "Description"

		directorApp := Application{
			ID:          "id1",
			Name:        "App1",
			Description: &description,
			Labels:      nil, // TODO: convert case with not nil
			APIs:        []APIDefinition{},
			EventAPIs:   []EventAPIDefinition{},
			Documents:   []Document{},
		}

		expected := v1alpha1.Application{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Application",
				APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "App1",
			},
			Spec: v1alpha1.ApplicationSpec{
				Description:      "Description",
				SkipInstallation: false,
				Services:         []v1alpha1.Service{},
				AccessLabel:      "App1",
				Labels:           map[string]string{},
			},
		}

		// when
		application := converter.Do(directorApp)

		// then
		assert.Equal(t, expected, application)
	})

	t.Run("should convert application with services containing APIs", func(t *testing.T) {
		// given
		mockNameResolver := &k8smocks.NameResolver{}
		converter := NewConverter(mockNameResolver)

		mockNameResolver.On("GetResourceName", "App1", "serviceId1").Return("resourceName1")
		mockNameResolver.On("GetResourceName", "App1", "serviceId2").Return("resourceName2")
		mockNameResolver.On("GetResourceName", "App1", "serviceId3").Return("resourceName3")

		mockNameResolver.On("GetGatewayUrl", "App1", "serviceId1").Return("application-gateway.kyma-integration.svc.cluster.local")
		mockNameResolver.On("GetGatewayUrl", "App1", "serviceId2").Return("application-gateway.kyma-integration.svc.cluster.local")
		mockNameResolver.On("GetGatewayUrl", "App1", "serviceId3").Return("application-gateway.kyma-integration.svc.cluster.local")

		mockNameResolver.On("GetCredentialsSecretName", "App1", "serviceId2").Return("credentialsSecretName2")
		mockNameResolver.On("GetRequestParamsSecretName", "App1", "serviceId2").Return("paramatersSecretName2")

		mockNameResolver.On("GetCredentialsSecretName", "App1", "serviceId3").Return("credentialsSecretName3")
		mockNameResolver.On("GetRequestParamsSecretName", "App1", "serviceId3").Return("paramatersSecretName3")

		description := "Description"

		directorApp := Application{
			ID:          "id1",
			Name:        "App1",
			Description: &description,
			Labels:      nil, // TODO? Figure out what to do with labels
			APIs: []APIDefinition{
				{
					ID:                "serviceId1",
					Name:              "serviceName1",
					Description:       "API 1 description",
					TargetUrl:         "www.example.com/1",
					RequestParameters: RequestParameters{},
				},
				{
					ID:          "serviceId2",
					Name:        "serviceName2",
					Description: "API 2 description",
					TargetUrl:   "www.example.com/2",
					RequestParameters: RequestParameters{
						Headers: &map[string][]string{
							"key": {"value"},
						},
					},
					Credentials: &Credentials{
						Basic: &Basic{
							Username: "admin",
							Password: "nimda",
						},
					},
				},
				{
					ID:          "serviceId3",
					Name:        "serviceName3",
					Description: "API 3 description",
					TargetUrl:   "www.example.com/3",
					RequestParameters: RequestParameters{
						QueryParameters: &map[string][]string{
							"key": {"value"},
						},
					},
					Credentials: &Credentials{
						Oauth: &Oauth{
							URL:          "www.oauth.com/3",
							ClientID:     "client_id",
							ClientSecret: "client_secret",
						},
						CSRFInfo: &CSRFInfo{
							TokenEndpointURL: "www.csrf.com/3",
						},
					},
				},
			},
			EventAPIs: []EventAPIDefinition{},
			Documents: []Document{},
		}

		expected := v1alpha1.Application{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Application",
				APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "App1",
			},
			Spec: v1alpha1.ApplicationSpec{
				Description:      "Description",
				SkipInstallation: false,
				AccessLabel:      "App1",
				Labels:           map[string]string{},
				Services: []v1alpha1.Service{
					{
						ID:          "serviceId1",
						Identifier:  "",
						Name:        "servicename1-cb830",
						DisplayName: "serviceName1",
						Description: "API 1 description",
						Labels: map[string]string{
							"connected-app": "App1",
						},
						LongDescription:     "",
						ProviderDisplayName: "",
						Tags:                []string{},
						Entries: []v1alpha1.Entry{
							{
								Type:                        specAPIType,
								GatewayUrl:                  "application-gateway.kyma-integration.svc.cluster.local",
								AccessLabel:                 "resourceName1",
								TargetUrl:                   "www.example.com/1",
								SpecificationUrl:            "",
								RequestParametersSecretName: "",
							},
						},
					},
					{
						ID:          "serviceId2",
						Identifier:  "",
						Name:        "servicename2-b25a8",
						DisplayName: "serviceName2",
						Description: "API 2 description",
						Labels: map[string]string{
							"connected-app": "App1",
						},
						LongDescription:     "",
						ProviderDisplayName: "",
						Tags:                []string{},
						Entries: []v1alpha1.Entry{
							{
								Type:             specAPIType,
								GatewayUrl:       "application-gateway.kyma-integration.svc.cluster.local",
								AccessLabel:      "resourceName2",
								TargetUrl:        "www.example.com/2",
								SpecificationUrl: "",
								Credentials: v1alpha1.Credentials{
									Type:              CredentialsBasicType,
									SecretName:        "credentialsSecretName2",
									AuthenticationUrl: "",
								},
								RequestParametersSecretName: "paramatersSecretName2",
							},
						},
					},
					{
						ID:          "serviceId3",
						Identifier:  "",
						Name:        "servicename3-1836d",
						DisplayName: "serviceName3",
						Description: "API 3 description",
						Labels: map[string]string{
							"connected-app": "App1",
						},
						LongDescription:     "",
						ProviderDisplayName: "",
						Tags:                []string{},
						Entries: []v1alpha1.Entry{
							{
								Type:             specAPIType,
								GatewayUrl:       "application-gateway.kyma-integration.svc.cluster.local",
								AccessLabel:      "resourceName3",
								TargetUrl:        "www.example.com/3",
								SpecificationUrl: "",
								Credentials: v1alpha1.Credentials{
									Type:              CredentialsOAuthType,
									SecretName:        "credentialsSecretName3",
									AuthenticationUrl: "www.oauth.com/3",
									CSRFInfo: &v1alpha1.CSRFInfo{
										TokenEndpointURL: "www.csrf.com/3",
									},
								},
								RequestParametersSecretName: "paramatersSecretName3",
							},
						},
					},
				},
			},
		}

		// when
		application := converter.Do(directorApp)

		// then
		assert.Equal(t, expected, application)
	})

	t.Run("should convert application with services containing events", func(t *testing.T) {
		// given
		mockNameResolver := &k8smocks.NameResolver{}
		converter := NewConverter(mockNameResolver)

		mockNameResolver.On("GetResourceName", "App1", "serviceId1").Return("resourceName1")

		description := "Description"

		directorApp := Application{
			ID:          "id1",
			Name:        "App1",
			Description: &description,
			Labels:      nil, // TODO: convert case with not nil
			APIs:        []APIDefinition{},
			EventAPIs: []EventAPIDefinition{
				{
					ID:          "serviceId1",
					Name:        "serviceName1",
					Description: "Events 1 description",
				},
			},
			Documents: []Document{},
		}

		expected := v1alpha1.Application{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Application",
				APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "App1",
			},
			Spec: v1alpha1.ApplicationSpec{
				Description:      "Description",
				SkipInstallation: false,
				Services: []v1alpha1.Service{
					{
						ID:          "serviceId1",
						Identifier:  "",
						Name:        "servicename1-cb830",
						DisplayName: "serviceName1",
						Description: "Events 1 description",
						Labels: map[string]string{
							"connected-app": "App1",
						},
						LongDescription:     "",
						ProviderDisplayName: "",
						Tags:                []string{},
						Entries: []v1alpha1.Entry{
							{
								Type:             specEventsType,
								AccessLabel:      "resourceName1",
								SpecificationUrl: "",
							},
						},
					},
				},
				AccessLabel: "App1",
				Labels:      map[string]string{},
			},
		}

		// when
		application := converter.Do(directorApp)

		// then
		assert.Equal(t, expected, application)
	})

	t.Run("should convert application with services containing API without credentials and events", func(t *testing.T) {

	})
}
