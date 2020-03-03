package applications

import (
	"testing"

	"kyma-project.io/compass-runtime-agent/internal/kyma/model"

	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8smocks "kyma-project.io/compass-runtime-agent/internal/k8sconsts/mocks"
)

func TestConverter(t *testing.T) {

	t.Run("should convert application without services", func(t *testing.T) {
		// given
		mockNameResolver := &k8smocks.NameResolver{}
		converter := NewConverter(mockNameResolver)

		directorApp := model.Application{
			ID:          "App1",
			Name:        "Appname1",
			Description: "Description",
			Labels: map[string]interface{}{
				"keySlice": []string{"value1", "value2"},
				"key":      "value",
			},
			APIs:           []model.APIDefinition{},
			EventAPIs:      []model.EventAPIDefinition{},
			Documents:      []model.Document{},
			SystemAuthsIDs: []string{"auth1", "auth2"},
		}

		expected := v1alpha1.Application{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Application",
				APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "Appname1",
			},
			Spec: v1alpha1.ApplicationSpec{
				Description:      "Description",
				SkipInstallation: false,
				Services:         []v1alpha1.Service{},
				AccessLabel:      "Appname1",
				Labels: map[string]string{
					"keySlice": "value1,value2",
					"key":      "value",
				},
				CompassMetadata: &v1alpha1.CompassMetadata{ApplicationID: "App1", Authentication: v1alpha1.Authentication{ClientIds: []string{"auth1", "auth2"}}},
			},
		}

		// when
		application := converter.Do(directorApp)

		// then
		assert.Equal(t, expected, application)
	})

	t.Run("should convert application with services containing protected APIs", func(t *testing.T) {
		// given
		mockNameResolver := &k8smocks.NameResolver{}
		converter := NewConverter(mockNameResolver)

		mockNameResolver.On("GetResourceName", "Appname1", "serviceId1").Return("resourceName1")
		mockNameResolver.On("GetResourceName", "Appname1", "serviceId2").Return("resourceName2")

		mockNameResolver.On("GetGatewayUrl", "Appname1", "serviceId1").Return("application-gateway.kyma-integration.svc.cluster.local")
		mockNameResolver.On("GetGatewayUrl", "Appname1", "serviceId2").Return("application-gateway.kyma-integration.svc.cluster.local")

		mockNameResolver.On("GetCredentialsSecretName", "Appname1", "serviceId1").Return("credentialsSecretName1")
		mockNameResolver.On("GetRequestParamsSecretName", "Appname1", "serviceId1").Return("paramatersSecretName1")

		mockNameResolver.On("GetCredentialsSecretName", "Appname1", "serviceId2").Return("credentialsSecretName2")
		mockNameResolver.On("GetRequestParamsSecretName", "Appname1", "serviceId2").Return("paramatersSecretName2")

		directorApp := model.Application{
			ID:                  "App1",
			Name:                "Appname1",
			Description:         "Description",
			ProviderDisplayName: "provider",
			Labels:              nil,
			APIs: []model.APIDefinition{
				{
					ID:          "serviceId1",
					Name:        "serviceName1",
					Description: "",
					TargetUrl:   "www.example.com/1",
					Auth: &model.Auth{
						RequestParameters: model.RequestParameters{
							Headers: &map[string][]string{
								"key": {"value"},
							},
						},
						Credentials: &model.Credentials{
							Basic: &model.Basic{
								Username: "admin",
								Password: "nimda",
							},
						},
					},
				},
				{
					ID:          "serviceId2",
					Name:        "serviceName2",
					Description: "API 2 description",
					TargetUrl:   "www.example.com/2",
					Auth: &model.Auth{
						RequestParameters: model.RequestParameters{
							QueryParameters: &map[string][]string{
								"key": {"value"},
							},
						},
						Credentials: &model.Credentials{
							Oauth: &model.Oauth{
								URL:          "www.oauth.com/2",
								ClientID:     "client_id",
								ClientSecret: "client_secret",
							},
							CSRFInfo: &model.CSRFInfo{
								TokenEndpointURL: "www.csrf.com/2",
							},
						},
					},
				},
			},
			EventAPIs:      []model.EventAPIDefinition{},
			Documents:      []model.Document{},
			SystemAuthsIDs: []string{"auth1", "auth2"},
		}

		expected := v1alpha1.Application{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Application",
				APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "Appname1",
			},
			Spec: v1alpha1.ApplicationSpec{
				Description:      "Description",
				SkipInstallation: false,
				AccessLabel:      "Appname1",
				Labels:           map[string]string{},
				CompassMetadata:  &v1alpha1.CompassMetadata{ApplicationID: "App1", Authentication: v1alpha1.Authentication{ClientIds: []string{"auth1", "auth2"}}},
				Services: []v1alpha1.Service{
					{
						ID:          "serviceId1",
						Identifier:  "",
						Name:        "servicename1-cb830",
						DisplayName: "serviceName1",
						Description: "Description not provided",
						Labels: map[string]string{
							"connected-app": "Appname1",
						},
						LongDescription:     "",
						ProviderDisplayName: "provider",
						Tags:                []string{},
						Entries: []v1alpha1.Entry{
							{
								Type:             SpecAPIType,
								GatewayUrl:       "application-gateway.kyma-integration.svc.cluster.local",
								AccessLabel:      "resourceName1",
								TargetUrl:        "www.example.com/1",
								SpecificationUrl: "",
								Credentials: v1alpha1.Credentials{
									Type:              CredentialsBasicType,
									SecretName:        "credentialsSecretName1",
									AuthenticationUrl: "",
								},
								RequestParametersSecretName: "paramatersSecretName1",
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
							"connected-app": "Appname1",
						},
						LongDescription:     "",
						ProviderDisplayName: "provider",
						Tags:                []string{},
						Entries: []v1alpha1.Entry{
							{
								Type:             SpecAPIType,
								GatewayUrl:       "application-gateway.kyma-integration.svc.cluster.local",
								AccessLabel:      "resourceName2",
								TargetUrl:        "www.example.com/2",
								SpecificationUrl: "",
								Credentials: v1alpha1.Credentials{
									Type:              CredentialsOAuthType,
									SecretName:        "credentialsSecretName2",
									AuthenticationUrl: "www.oauth.com/2",
									CSRFInfo: &v1alpha1.CSRFInfo{
										TokenEndpointURL: "www.csrf.com/2",
									},
								},
								RequestParametersSecretName: "paramatersSecretName2",
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

	t.Run("should convert application with services containing events and API, and no System Auths", func(t *testing.T) {
		// given
		mockNameResolver := &k8smocks.NameResolver{}
		converter := NewConverter(mockNameResolver)

		mockNameResolver.On("GetResourceName", "Appname1", "serviceId1").Return("resourceName1")
		mockNameResolver.On("GetResourceName", "Appname1", "serviceId2").Return("resourceName2")

		mockNameResolver.On("GetGatewayUrl", "Appname1", "serviceId1").Return("application-gateway.kyma-integration.svc.cluster.local")

		directorApp := model.Application{
			ID:                  "App1",
			Name:                "Appname1",
			Description:         "Description",
			ProviderDisplayName: "provider",
			Labels:              nil,
			APIs: []model.APIDefinition{
				{
					ID:          "serviceId1",
					Name:        "veryveryveryveryveryveryveryveryveryveryveryveryveryveryveryveryveryveryveryverylongserviceName1",
					Description: "API 1 description",
					TargetUrl:   "www.example.com/1",
					APISpec: &model.APISpec{
						Type: model.APISpecTypeOpenAPI,
					},
					Auth: &model.Auth{
						RequestParameters: model.RequestParameters{},
					},
				},
			},
			EventAPIs: []model.EventAPIDefinition{
				{
					ID:          "serviceId2",
					Name:        "serviceName2",
					Description: "Events 1 description",
				},
			},
			Documents: []model.Document{},
		}

		expected := v1alpha1.Application{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Application",
				APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "Appname1",
			},
			Spec: v1alpha1.ApplicationSpec{
				Description:      "Description",
				SkipInstallation: false,
				CompassMetadata:  &v1alpha1.CompassMetadata{ApplicationID: "App1", Authentication: v1alpha1.Authentication{ClientIds: nil}},
				Services: []v1alpha1.Service{
					{
						ID:          "serviceId1",
						Identifier:  "",
						Name:        "veryveryveryveryveryveryveryveryveryveryveryveryveryveryv-cb830",
						DisplayName: "veryveryveryveryveryveryveryveryveryveryveryveryveryveryveryveryveryveryveryverylongserviceName1",
						Description: "API 1 description",
						Labels: map[string]string{
							"connected-app": "Appname1",
						},
						LongDescription:     "",
						ProviderDisplayName: "provider",
						Tags:                []string{},
						Entries: []v1alpha1.Entry{
							{
								Type:                        SpecAPIType,
								GatewayUrl:                  "application-gateway.kyma-integration.svc.cluster.local",
								AccessLabel:                 "resourceName1",
								TargetUrl:                   "www.example.com/1",
								ApiType:                     string(model.APISpecTypeOpenAPI),
								SpecificationUrl:            "",
								RequestParametersSecretName: "",
								Credentials:                 v1alpha1.Credentials{},
							},
						},
					},
					{
						ID:          "serviceId2",
						Identifier:  "",
						Name:        "servicename2-b25a8",
						DisplayName: "serviceName2",
						Description: "Events 1 description",
						Labels: map[string]string{
							"connected-app": "Appname1",
						},
						LongDescription:     "",
						ProviderDisplayName: "provider",
						Tags:                []string{},
						Entries: []v1alpha1.Entry{
							{
								Type:             SpecEventsType,
								AccessLabel:      "resourceName2",
								SpecificationUrl: "",
								Credentials:      v1alpha1.Credentials{},
							},
						},
					},
				},
				AccessLabel: "Appname1",
				Labels:      map[string]string{},
			},
		}

		// when
		application := converter.Do(directorApp)

		// then
		assert.Equal(t, expected, application)
	})
}
