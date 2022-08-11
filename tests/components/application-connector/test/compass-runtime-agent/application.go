package compass_runtime_agent

import (
	"bytes"
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"reflect"
)

type SecretComparer interface {
	Do(secretType, actual, expected string) bool
}

type Secret struct{}

func (s Secret) Do(secretType, actual, expected string) bool {
	return true
}


DisplayName:         "applicationOneDisplay",
ProviderDisplayName: "applicationOneDisplay",
LongDescription:     "applicationOne Test",
SkipVerify:          true,

func Compare(applicationCRD *v1alpha1.Application, applicationCRD_expected *v1alpha1.Application, comparer SecretComparer) bool {

	nameAndNamespaceEqual := applicationCRD.Name == applicationCRD_expected.Name && applicationCRD.Namespace == applicationCRD_expected.Namespace
	descriptionEqual := applicationCRD.Spec.Description == applicationCRD_expected.Spec.Description
	skipInstalationEqual := applicationCRD.Spec.SkipInstallation == applicationCRD_expected.Spec.SkipInstallation
	servicesEqual := compareServices(applicationCRD.Spec.Services, applicationCRD_expected.Spec.Services)
	labelsEqual := compareLabels(applicationCRD.Spec.Labels,applicationCRD_expected.Spec.Labels)
	tenantEqual := applicationCRD.Spec.Tenant == applicationCRD_expected.Spec.Tenant
	groupEqual := applicationCRD.Spec.Group == applicationCRD_expected.Spec.Group
	compassMetadataEqual := compareCompassMetadata(applicationCRD.Spec.CompassMetadata, applicationCRD_expected.Spec.CompassMetadata)
	tagsEqual := compareTags(applicationCRD.Spec.Tags, applicationCRD_expected.Spec.Tags)
	displayNameEqual := applicationCRD.Spec.DisplayName == applicationCRD_expected.Spec.DisplayName
	providerDisplayNameEqual := applicationCRD.Spec.ProviderDisplayName == applicationCRD_expected.Spec.ProviderDisplayName
	longDescriptionEqual := applicationCRD.Spec.LongDescription == applicationCRD_expected.Spec.LongDescription
	skipVerifyEqual := applicationCRD.Spec.SkipVerify == applicationCRD_expected.Spec.SkipVerify
	return nameAndNamespaceEqual && descriptionEqual && compassMetadataEqual && servicesEqual
}

func compareLabels(actual, expected map[string]string) bool {
	if len(actual) != len(expected) {
		return false
	}

	return true
}

func compareTags(actual, expected []string) bool {
	if len(actual) != len(expected) {
		return false
	}

	return true
}

func compareCompassMetadata(actual, expected *v1alpha1.CompassMetadata) bool {

	applicationIdEqual := actual.ApplicationID == expected.ApplicationID
	authenticationEqual := compareAuthentication(actual.Authentication, expected.Authentication)

	return applicationIdEqual && authenticationEqual
}

func compareAuthentication (actual, expected v1alpha1.Authentication) bool {
	if len(actual.ClientIds) != len(expected.ClientIds) {
		return false
	}

	return reflect.DeepEqual(actual.ClientIds,expected.ClientIds)
}

func compareServices(actual, expected []v1alpha1.Service) bool {

	if len(actual) != len(expected) {
		return false
	}

	for _, serviceActual := range actual {
		isFound := false
		for _, serviceExpected := range expected {
			nameEqual := serviceExpected.Name == serviceActual.Name
			idEqual := serviceExpected.ID == serviceActual.ID
			identifierEqual := serviceExpected.Identifier == serviceActual.Identifier
			displayNameEqual := serviceExpected.DisplayName == serviceActual.DisplayName
			descriptionEqual := serviceExpected.Description == serviceActual.Description
			entriesEqual := compareEntries(serviceActual.Entries, serviceExpected.Entries)
			authParameterSchemaEqual := serviceExpected.AuthCreateParameterSchema == serviceActual.AuthCreateParameterSchema
			isFound = nameEqual && idEqual && identifierEqual && displayNameEqual && descriptionEqual && entriesEqual && authParameterSchemaEqual
		}
		if !isFound {
			return false
		}
	}
	return true
}

func compareEntries(actual, expected []v1alpha1.Entry) bool {

	if len(actual) != len(expected) {
		return false
	}

	for _, entryActual := range actual {
		isFound := false
		for _, entryExpected := range expected {
			typeEqual := entryActual.Type == entryExpected.Type
			targetUrlEqual := entryActual.TargetUrl == entryExpected.TargetUrl
			specificationUrlEqual := entryActual.SpecificationUrl == entryExpected.SpecificationUrl
			apiTypeEqual := entryActual.ApiType == entryExpected.ApiType
			credentialsEqual := compareCredentials(entryActual.Credentials, entryExpected.Credentials)
			requestParameterSecretEqual := entryActual.RequestParametersSecretName == entryExpected.RequestParametersSecretName
			nameEqual := entryActual.Name == entryExpected.Name
			idEqual := entryActual.ID == entryExpected.ID
			centralGatewayUrlEqual := entryActual.CentralGatewayUrl == entryExpected.CentralGatewayUrl

			isFound = typeEqual && targetUrlEqual && specificationUrlEqual && apiTypeEqual && credentialsEqual && requestParameterSecretEqual && nameEqual && idEqual && centralGatewayUrlEqual
		}
		if !isFound {
			return false
		}
	}
	return true
}

//TODO compare whole credentials data, not only .yaml name
func compareCredentials(actual, expected v1alpha1.Credentials) bool {

	if actual == (v1alpha1.Credentials{}) && expected != (v1alpha1.Credentials{}) || expected == (v1alpha1.Credentials{}) && actual != (v1alpha1.Credentials{}) {
		return false
	}

	typeEqual := actual.Type == expected.Type
	secretNameEqual := actual.SecretName == expected.SecretName
	authenticationUrlEqual := actual.AuthenticationUrl == expected.AuthenticationUrl
	csrfInfoEqual := compareCSRF(actual.CSRFInfo, expected.CSRFInfo)

	return typeEqual && secretNameEqual && authenticationUrlEqual && csrfInfoEqual
}

func compareCSRF(actual, expected *v1alpha1.CSRFInfo) bool {

	if actual == nil && expected != nil || actual != nil && expected == nil {
		return false
	}

	return actual.TokenEndpointURL == expected.TokenEndpointURL
}

//future
func compareSpec(actual, expected v1alpha1.ApplicationSpec) bool {
	return false
}
