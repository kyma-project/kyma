package converters

import (
	"strings"

	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kyma-project.io/compass-runtime-agent/internal/k8sconsts"
	"kyma-project.io/compass-runtime-agent/internal/kyma/model"
)

type apiPackageConverter struct {
	nameResolver k8sconsts.NameResolver
}

func (c apiPackageConverter) Do(application model.Application) v1alpha1.Application {

	convertLabels := func(directorLabels model.Labels) map[string]string {
		labels := make(map[string]string)

		for key, value := range directorLabels {
			switch value.(type) {
			case string:
				labels[key] = value.(string)
				break
			case []string:
				newVal := strings.Join(value.([]string), ",")
				labels[key] = newVal
				break
			}
		}

		return labels
	}

	return v1alpha1.Application{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Application",
			APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: application.Name,
		},
		Spec: v1alpha1.ApplicationSpec{
			Description:      application.Description,
			SkipInstallation: false,
			Labels:           convertLabels(application.Labels),
			Services:         c.toServices(application.Name, application.ProviderDisplayName, application.APIPackages),
			CompassMetadata:  c.toCompassMetadata(application.ID, application.SystemAuthsIDs),
		},
	}
}

func (c apiPackageConverter) toServices(applicationName, appProvider string, packages []model.APIPackage) []v1alpha1.Service {
	services := make([]v1alpha1.Service, 0, len(packages))

	for _, p := range packages {
		c.toAPIService(applicationName, appProvider, p)
	}

	return services
}

func (c apiPackageConverter) toAPIService(applicationName, appProvider string, apiPackage model.APIPackage) v1alpha1.Service {

	description := *apiPackage.Description
	if description == "" {
		description = "Description not provided"
	}

	return v1alpha1.Service{
		ID:          apiPackage.ID,
		Identifier:  "", // not available in the Director's API
		Name:        createServiceName(apiPackage.Name, apiPackage.ID),
		DisplayName: apiPackage.Name,
		Description: description,
		Entries:     c.toServiceEntries(apiPackage.APIDefinitions, apiPackage.EventDefinitions),
	}
}

func (c apiPackageConverter) toServiceEntries(apiDefinitions []model.APIDefinition, eventAPIDefinitions []model.EventAPIDefinition) []v1alpha1.Entry {
	entries := make([]v1alpha1.Entry, 0, len(apiDefinitions)+len(eventAPIDefinitions))

	for _, apiDefinition := range apiDefinitions {
		entries = append(entries, c.toAPIEntry(apiDefinition))
	}

	for _, eventAPIDefinition := range eventAPIDefinitions {
		entries = append(entries, c.toEventServiceEntry(eventAPIDefinition))
	}

	return entries
}

func (c apiPackageConverter) toAPIEntry(apiDefinition model.APIDefinition) v1alpha1.Entry {

	getApiType := func() string {
		if apiDefinition.APISpec != nil {
			return string(apiDefinition.APISpec.Type)
		}

		return ""
	}

	entry := v1alpha1.Entry{
		Type:             SpecAPIType,
		ApiType:          getApiType(),
		TargetUrl:        apiDefinition.TargetUrl,
		SpecificationUrl: "", // Director returns BLOB here
	}

	return entry
}

func (c apiPackageConverter) toEventServiceEntry(eventsDefinition model.EventAPIDefinition) v1alpha1.Entry {
	return v1alpha1.Entry{
		Type:             SpecEventsType,
		SpecificationUrl: "", // Director returns BLOB here
	}
}

func (c apiPackageConverter) toCompassMetadata(applicationID string, systemAuthsIDs []string) *v1alpha1.CompassMetadata {
	return &v1alpha1.CompassMetadata{
		ApplicationID: applicationID,
		Authentication: v1alpha1.Authentication{
			ClientIds: systemAuthsIDs,
		},
	}
}
