package converters

import (
	"testing"

	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kyma-project.io/compass-runtime-agent/internal/kyma/model"
)

func TestNsConverter(t *testing.T) {

	t.Run("should convert application without API packages", func(t *testing.T) {
		// given
		converter := NewGatewayForNsConverter()

		directorApp := model.Application{
			ID:   "App1",
			Name: "Appname1",
			Labels: map[string]interface{}{
				"keySlice": []string{"value1", "value2"},
				"key":      "value",
			},
			APIPackages:    []model.APIPackage{},
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
				Description:      "Description not provided",
				SkipInstallation: false,
				Services:         []v1alpha1.Service{},
				Labels: map[string]string{
					"keySlice":   "value1,value2",
					"key":        "value",
					connectedApp: "Appname1",
				},
				CompassMetadata: &v1alpha1.CompassMetadata{ApplicationID: "App1", Authentication: v1alpha1.Authentication{ClientIds: []string{"auth1", "auth2"}}},
			},
		}

		// when
		application := converter.Do(directorApp)

		// then
		assert.Equal(t, expected, application)
	})

	t.Run("should convert application containing API Packages with API Definitions", func(t *testing.T) {
		// given
		converter := NewGatewayForNsConverter()
		instanceAuthRequestInputSchema := "{}"

		description := "description"
		directorApp := model.Application{
			ID:                  "App1",
			Name:                "Appname1",
			Description:         "Description",
			ProviderDisplayName: "provider",
			Labels:              nil,
			APIPackages: []model.APIPackage{
				{
					ID:                             "package1",
					Name:                           "packageName1",
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
				},
				{
					ID:          "package2",
					Name:        "packageName2",
					Description: &description,
					APIDefinitions: []model.APIDefinition{
						{
							ID:          "serviceId3",
							Name:        "serviceName3",
							Description: "",
							TargetUrl:   "www.example.com/3",
						},
					},
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
				Name: "Appname1",
			},
			Spec: v1alpha1.ApplicationSpec{
				Description:      "Description",
				SkipInstallation: false,
				Labels: map[string]string{
					connectedApp: "Appname1",
				},
				CompassMetadata: &v1alpha1.CompassMetadata{ApplicationID: "App1", Authentication: v1alpha1.Authentication{ClientIds: []string{"auth1", "auth2"}}},
				Services: []v1alpha1.Service{
					{
						ID:                        "package1",
						Identifier:                "",
						Name:                      "packagename1-8996f",
						DisplayName:               "packageName1",
						Description:               "Description not provided",
						AuthCreateParameterSchema: &instanceAuthRequestInputSchema,
						Entries: []v1alpha1.Entry{
							{
								ID:        "serviceId1",
								Name:      "serviceName1",
								Type:      SpecAPIType,
								TargetUrl: "www.example.com/1",
							},
							{
								ID:        "serviceId2",
								Name:      "serviceName2",
								Type:      SpecAPIType,
								TargetUrl: "www.example.com/2",
							},
						},
					},
					{
						ID:          "package2",
						Identifier:  "",
						Name:        "packagename2-60248",
						DisplayName: "packageName2",
						Description: "description",
						Entries: []v1alpha1.Entry{
							{
								ID:        "serviceId3",
								Name:      "serviceName3",
								Type:      SpecAPIType,
								TargetUrl: "www.example.com/3",
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
		converter := NewGatewayForNsConverter()

		directorApp := model.Application{
			ID:                  "App1",
			Name:                "Appname1",
			Description:         "Description",
			ProviderDisplayName: "provider",
			Labels:              nil,
			APIPackages: []model.APIPackage{
				{
					ID:   "package1",
					Name: "packageName1",
					APIDefinitions: []model.APIDefinition{
						{
							ID:          "serviceId1",
							Name:        "veryveryveryveryveryveryveryveryveryveryveryveryveryveryveryveryveryveryveryverylongserviceName1",
							Description: "API 1 description",
							TargetUrl:   "www.example.com/1",
							APISpec: &model.APISpec{
								Type: model.APISpecTypeOpenAPI,
							},
							RequestParameters: model.RequestParameters{},
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
				Name: "Appname1",
			},
			Spec: v1alpha1.ApplicationSpec{
				Description:      "Description",
				SkipInstallation: false,
				Labels: map[string]string{
					connectedApp: "Appname1",
				},
				CompassMetadata: &v1alpha1.CompassMetadata{ApplicationID: "App1", Authentication: v1alpha1.Authentication{ClientIds: nil}},
				Services: []v1alpha1.Service{
					{
						ID:          "package1",
						Identifier:  "",
						Name:        "packagename1-8996f",
						DisplayName: "packageName1",
						Description: "Description not provided",
						Entries: []v1alpha1.Entry{
							{
								ID:        "serviceId1",
								Name:      "veryveryveryveryveryveryveryveryveryveryveryveryveryveryveryveryveryveryveryverylongserviceName1",
								Type:      SpecAPIType,
								TargetUrl: "www.example.com/1",
								ApiType:   string(model.APISpecTypeOpenAPI),
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
