package compass_runtime_agent

import (
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

func Compare(applicationCRD *v1alpha1.Application, applicationCRD_expected *v1alpha1.Application, comparer SecretComparer) bool {

	specEqual := compareSpec(applicationCRD.Spec, applicationCRD_expected.Spec)
	nameAndNamespaceEqual := applicationCRD.Name == applicationCRD_expected.Name && applicationCRD.Namespace == applicationCRD_expected.Namespace

	return nameAndNamespaceEqual && specEqual
}

func compareLabels(actual, expected map[string]string) bool {
	if len(actual) != len(expected) {
		return false
	}

	return reflect.DeepEqual(actual, expected)
}

func compareTags(actual, expected []string) bool {
	if len(actual) != len(expected) {
		return false
	}

	return reflect.DeepEqual(actual, expected)
}

func compareCompassMetadata(actual, expected *v1alpha1.CompassMetadata) bool {

	applicationIdEqual := actual.ApplicationID == expected.ApplicationID
	authenticationEqual := compareAuthentication(actual.Authentication, expected.Authentication)

	return applicationIdEqual && authenticationEqual
}

func compareAuthentication(actual, expected v1alpha1.Authentication) bool {
	if len(actual.ClientIds) != len(expected.ClientIds) {
		return false
	}

	return reflect.DeepEqual(actual.ClientIds, expected.ClientIds)
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

func compareSpec(actual, expected v1alpha1.ApplicationSpec) bool {

	descriptionEqual := actual.Description == expected.Description
	skipInstallationEqual := actual.SkipInstallation == expected.SkipInstallation
	servicesEqual := compareServices(actual.Services, expected.Services)
	labelsEqual := compareLabels(actual.Labels, expected.Labels)
	tenantEqual := actual.Tenant == expected.Tenant
	groupEqual := actual.Group == expected.Group
	compassMetadataEqual := compareCompassMetadata(actual.CompassMetadata, expected.CompassMetadata)
	tagsEqual := compareTags(actual.Tags, expected.Tags)
	displayNameEqual := actual.DisplayName == expected.DisplayName
	providerDisplayNameEqual := actual.ProviderDisplayName == expected.ProviderDisplayName
	longDescriptionEqual := actual.LongDescription == expected.LongDescription
	skipVerifyEqual := actual.SkipVerify == expected.SkipVerify

	return descriptionEqual && compassMetadataEqual && skipInstallationEqual &&
		servicesEqual && labelsEqual && tenantEqual && groupEqual && tagsEqual && displayNameEqual && providerDisplayNameEqual &&
		longDescriptionEqual && skipVerifyEqual
}
