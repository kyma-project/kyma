package applications

import (
	"testing"

	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/k8sconsts"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/model"
)

const (
	centralGatewayServiceUrl = "http://central-application-gateway.kyma-system.svc.cluster.local:8082"
)

func TestConverter(t *testing.T) {
	t.Run("should convert application without API bundles", func(t *testing.T) {
		// given
		converter := NewConverter(k8sconsts.NewNameResolver(), centralGatewayServiceUrl, false)

		directorApp := model.Application{
			ID:   "App1",
			Name: "Appname1",
			Labels: map[string]interface{}{
				"keySlice": []string{"value1", "value2"},
				"key":      "value",
			},
			ApiBundles:     []model.APIBundle{},
			SystemAuthsIDs: []string{"auth1", "auth2"},
		}

		expected := v1alpha1.Application{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Application",
				APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:   "Appname1",
				Labels: map[string]string{managedByLabelKey: managedByLabelValue},
			},
			Spec: v1alpha1.ApplicationSpec{
				Description:      "Description not provided",
				SkipInstallation: false,
				Services:         []v1alpha1.Service{},
				Labels: map[string]string{
					connectedAppLabelKey: "Appname1",
				},
				CompassMetadata: &v1alpha1.CompassMetadata{ApplicationID: "App1", Authentication: v1alpha1.Authentication{ClientIds: []string{"auth1", "auth2"}}},
			},
		}

		// when
		application := converter.Do(directorApp)

		// then
		assert.Equal(t, expected, application)
	})

	t.Run("should convert application containing API Bundles with API Definitions", func(t *testing.T) {
		// given
		converter := NewConverter(k8sconsts.NewNameResolver(), centralGatewayServiceUrl, false)
		instanceAuthRequestInputSchema := "{}"

		emptyDescription := ""
		description := "description"
		directorApp := model.Application{
			ID:                  "App1",
			Name:                "Appname1",
			Description:         "Description",
			ProviderDisplayName: "provider",
			Labels:              nil,
			ApiBundles: []model.APIBundle{
				{
					ID:                             "bundle1",
					Name:                           "bundleName1",
					InstanceAuthRequestInputSchema: &instanceAuthRequestInputSchema,
					APIDefinitions: []model.APIDefinition{
						{
							ID:          "serviceId1",
							Name:        "serviceName1",
							Description: "",
							TargetUrl:   "www.example.com/1",
						},
						{
							ID:          "serviceId2",
							Name:        "serviceName2",
							Description: "API 2 description",
							TargetUrl:   "www.example.com/2",
						},
					},
					DefaultInstanceAuth: &model.Auth{
						Credentials: &model.Credentials{
							Oauth: &model.Oauth{
								URL:          "https://oauth.example.com",
								ClientID:     "test-client",
								ClientSecret: "test-secret",
							},
							CSRFInfo: &model.CSRFInfo{
								TokenEndpointURL: "https://tokern.example.com",
							},
						},
					},
				},
				{
					ID:          "bundle2",
					Name:        "bundleName2",
					Description: &description,
					APIDefinitions: []model.APIDefinition{
						{
							ID:          "serviceId3",
							Name:        "serviceName3",
							Description: "",
							TargetUrl:   "www.example.com/3",
						},
					},
					DefaultInstanceAuth: &model.Auth{
						Credentials: &model.Credentials{
							Basic: &model.Basic{
								Username: "my-username",
								Password: "my-password",
							},
						},
						RequestParameters: &model.RequestParameters{
							Headers:         &map[string][]string{"header": {"header-value"}},
							QueryParameters: &map[string][]string{"query-param": {"query-param-value"}},
						},
					},
				},
				{
					ID:             "bundle3",
					Name:           "bundleName3",
					Description:    &emptyDescription,
					APIDefinitions: []model.APIDefinition{},
				},
			},
			SystemAuthsIDs: []string{"auth1", "auth2"},
		}

		expected := v1alpha1.Application{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Application",
				APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:   "Appname1",
				Labels: map[string]string{managedByLabelKey: managedByLabelValue},
			},
			Spec: v1alpha1.ApplicationSpec{
				Description:      "Description",
				SkipInstallation: false,
				Labels: map[string]string{
					connectedAppLabelKey: "Appname1",
				},
				CompassMetadata: &v1alpha1.CompassMetadata{ApplicationID: "App1", Authentication: v1alpha1.Authentication{ClientIds: []string{"auth1", "auth2"}}},
				Services: []v1alpha1.Service{
					{
						ID:                        "bundle1",
						Identifier:                "",
						Name:                      "bundlename1-43857",
						DisplayName:               "bundleName1",
						Description:               "Description not provided",
						AuthCreateParameterSchema: &instanceAuthRequestInputSchema,
						Entries: []v1alpha1.Entry{
							{
								ID:                "serviceId1",
								Name:              "serviceName1",
								Type:              SpecAPIType,
								TargetUrl:         "www.example.com/1",
								CentralGatewayUrl: "http://central-application-gateway.kyma-system.svc.cluster.local:8082/Appname1/bundlename1/servicename1",
								Credentials: v1alpha1.Credentials{
									Type:              "OAuth",
									SecretName:        "Appname1-bundle1",
									AuthenticationUrl: "https://oauth.example.com",
									CSRFInfo: &v1alpha1.CSRFInfo{
										TokenEndpointURL: "https://tokern.example.com",
									},
								},
							},
							{
								ID:                "serviceId2",
								Name:              "serviceName2",
								Type:              SpecAPIType,
								TargetUrl:         "www.example.com/2",
								CentralGatewayUrl: "http://central-application-gateway.kyma-system.svc.cluster.local:8082/Appname1/bundlename1/servicename2",
								Credentials: v1alpha1.Credentials{
									Type:              "OAuth",
									SecretName:        "Appname1-bundle1",
									AuthenticationUrl: "https://oauth.example.com",
									CSRFInfo: &v1alpha1.CSRFInfo{
										TokenEndpointURL: "https://tokern.example.com",
									},
								},
							},
						},
					},
					{
						ID:          "bundle2",
						Identifier:  "",
						Name:        "bundlename2-4b91a",
						DisplayName: "bundleName2",
						Description: "description",
						Entries: []v1alpha1.Entry{
							{
								ID:                "serviceId3",
								Name:              "serviceName3",
								Type:              SpecAPIType,
								TargetUrl:         "www.example.com/3",
								CentralGatewayUrl: "http://central-application-gateway.kyma-system.svc.cluster.local:8082/Appname1/bundlename2/servicename3",
								Credentials: v1alpha1.Credentials{
									Type:       "Basic",
									SecretName: "Appname1-bundle2",
								},
								RequestParametersSecretName: "params-Appname1-bundle2",
							},
						},
					},
					{
						ID:          "bundle3",
						Identifier:  "",
						Name:        "bundlename3-16aa4",
						DisplayName: "bundleName3",
						Description: "Description not provided",
						Entries:     []v1alpha1.Entry{},
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
		converter := NewConverter(k8sconsts.NewNameResolver(), centralGatewayServiceUrl, false)

		directorApp := model.Application{
			ID:                  "App1",
			Name:                "Appname1",
			Description:         "Description",
			ProviderDisplayName: "provider",
			Labels:              nil,
			ApiBundles: []model.APIBundle{
				{
					ID:   "bundle1",
					Name: "bundleName1",
					APIDefinitions: []model.APIDefinition{
						{
							ID:          "serviceId1",
							Name:        "veryveryveryveryveryveryveryveryveryveryveryveryveryveryveryveryveryveryveryverylongserviceName1",
							Description: "API 1 description",
							TargetUrl:   "www.example.com/1",
						},
					},
					EventDefinitions: []model.EventAPIDefinition{
						{
							ID:          "serviceId2",
							Name:        "serviceName2",
							Description: "Events 1 description",
						},
					},
				},
			},
		}

		expected := v1alpha1.Application{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Application",
				APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:   "Appname1",
				Labels: map[string]string{managedByLabelKey: managedByLabelValue},
			},
			Spec: v1alpha1.ApplicationSpec{
				Description:      "Description",
				SkipInstallation: false,
				Labels: map[string]string{
					connectedAppLabelKey: "Appname1",
				},
				CompassMetadata: &v1alpha1.CompassMetadata{ApplicationID: "App1", Authentication: v1alpha1.Authentication{ClientIds: nil}},
				Services: []v1alpha1.Service{
					{
						ID:          "bundle1",
						Identifier:  "",
						Name:        "bundlename1-43857",
						DisplayName: "bundleName1",
						Description: "Description not provided",
						Entries: []v1alpha1.Entry{
							{
								ID:                "serviceId1",
								Name:              "veryveryveryveryveryveryveryveryveryveryveryveryveryveryveryveryveryveryveryverylongserviceName1",
								Type:              SpecAPIType,
								TargetUrl:         "www.example.com/1",
								CentralGatewayUrl: "http://central-application-gateway.kyma-system.svc.cluster.local:8082/Appname1/bundlename1/veryveryveryveryveryveryveryveryveryveryveryveryveryveryv",
							},
							{
								ID:   "serviceId2",
								Name: "serviceName2",
								Type: SpecEventsType,
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
}
