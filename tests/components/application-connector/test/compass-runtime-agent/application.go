package compass_runtime_agent

import "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"

type SecretComparer interface {
	Do(secretType, actual, expected string) bool
}

type Secret struct{}

func (s Secret) Do(secretType, actual, expected string) bool {
	return true
}

SkipInstallation: false,
Services:         services,
Labels:           nil,
Tenant:           "test",
Group:            "test",
CompassMetadata: &v1alpha1.CompassMetadata{
ApplicationID:  "compassID1",
Authentication: v1alpha1.Authentication{ClientIds: []string{"11", "22"}},
},
Tags:                []string{"tag1", "tag2"},
DisplayName:         "applicationOneDisplay",
ProviderDisplayName: "applicationOneDisplay",
LongDescription:     "applicationOne Test",
SkipVerify:          true,

func Compare(applicationCRD *v1alpha1.Application, applicationCRD_expected *v1alpha1.Application, comparer SecretComparer) bool {

	nameAndNamespaceEqual := applicationCRD.Name == applicationCRD_expected.Name && applicationCRD.Namespace == applicationCRD_expected.Namespace
	descriptionEqual := applicationCRD.Spec.Description == applicationCRD_expected.Spec.Description
	compassMetadataEqual := compareCompassMetadata(applicationCRD.Spec.CompassMetadata, applicationCRD_expected.Spec.CompassMetadata)
	servicesEqual := compareServices(applicationCRD.Spec.Services, applicationCRD_expected.Spec.Services)

	return nameAndNamespaceEqual && descriptionEqual && compassMetadataEqual && servicesEqual
}

func compareCompassMetadata(actual, expected *v1alpha1.CompassMetadata) bool {
	return true
}

func compareServices(actual, expected []v1alpha1.Service) bool {

	if len(actual) != len(expected) {
		return false
	}

	for _, serviceActual := range actual {
		isFound := false
		for _, servivceExpected := range expected{
			servivceExpected.
			isFound = true
		}
		if !isFound {
			return false
		}
	}
}

func compareEntries(actual, expected v1alpha1.Entry) bool {
	return true
}

//future
func compareSpec(actual, expected v1alpha1.ApplicationSpec) bool {
	return false
}