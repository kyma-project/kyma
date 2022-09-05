package compass_runtime_agent

import (
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"k8s.io/client-go/kubernetes"
	"reflect"
)

const actualSecretNamespace string = "kyma-integration"

func Compare(applicationCRD *v1alpha1.Application, applicationCRD_expected *v1alpha1.Application, comparer SecretClient, cli kubernetes.Interface) bool {

	specEqual := compareSpec(applicationCRD.Spec, applicationCRD_expected.Spec, comparer, cli)
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

func compareServices(actual, expected []v1alpha1.Service, comparer SecretClient, cli kubernetes.Interface) bool {

	if len(actual) != len(expected) {
		return false
	}

	for i := 0; i < len(actual); i++ {
		nameEqual := expected[i].Name == actual[i].Name
		idEqual := expected[i].ID == actual[i].ID
		identifierEqual := expected[i].Identifier == actual[i].Identifier
		displayNameEqual := expected[i].DisplayName == actual[i].DisplayName
		descriptionEqual := expected[i].Description == actual[i].Description
		entriesEqual := compareEntries(actual[i].Entries, expected[i].Entries, comparer, cli)
		authParameterSchemaEqual := expected[i].AuthCreateParameterSchema == actual[i].AuthCreateParameterSchema

		if !(nameEqual && idEqual && identifierEqual && displayNameEqual && descriptionEqual && entriesEqual && authParameterSchemaEqual) {
			return false
		}
	}
	return true
}

func compareEntries(actual, expected []v1alpha1.Entry, comparer SecretClient, cli kubernetes.Interface) bool {

	if len(actual) != len(expected) {
		return false
	}

	for i := 0; i < len(actual); i++ {
		typeEqual := actual[i].Type == actual[i].Type
		targetUrlEqual := actual[i].TargetUrl == actual[i].TargetUrl
		specificationUrlEqual := actual[i].SpecificationUrl == actual[i].SpecificationUrl
		apiTypeEqual := actual[i].ApiType == actual[i].ApiType
		credentialsEqual := compareCredentials(actual[i].Credentials, actual[i].Credentials, comparer, cli)
		requestParameterSecretEqual := actual[i].RequestParametersSecretName == actual[i].RequestParametersSecretName
		nameEqual := actual[i].Name == actual[i].Name
		idEqual := actual[i].ID == actual[i].ID
		centralGatewayUrlEqual := actual[i].CentralGatewayUrl == actual[i].CentralGatewayUrl

		if !(typeEqual && targetUrlEqual && specificationUrlEqual && apiTypeEqual && credentialsEqual && requestParameterSecretEqual && nameEqual && idEqual && centralGatewayUrlEqual) {
			return false
		}
	}
	return true
}

// TODO compare whole credentials data, not only .yaml name
func compareCredentials(actual, expected v1alpha1.Credentials, comparer SecretClient, cli kubernetes.Interface) bool {

	if actual == (v1alpha1.Credentials{}) && expected != (v1alpha1.Credentials{}) || expected == (v1alpha1.Credentials{}) && actual != (v1alpha1.Credentials{}) {
		return false
	}

	typeEqual := actual.Type == expected.Type

	if !typeEqual {
		return false
	}

	dataEqual := comparer.Compare(actual.SecretName, "oauth-test-expected", cli)
	secretNameEqual := actual.SecretName == expected.SecretName
	authenticationUrlEqual := actual.AuthenticationUrl == expected.AuthenticationUrl
	csrfInfoEqual := compareCSRF(actual.CSRFInfo, expected.CSRFInfo)
	return typeEqual && secretNameEqual && authenticationUrlEqual && csrfInfoEqual && dataEqual
}

func compareCSRF(actual, expected *v1alpha1.CSRFInfo) bool {

	if actual == nil && expected != nil || actual != nil && expected == nil {
		return false
	}

	return actual.TokenEndpointURL == expected.TokenEndpointURL
}

func compareSpec(actual, expected v1alpha1.ApplicationSpec, comparer SecretClient, cli kubernetes.Interface) bool {
	descriptionEqual := actual.Description == expected.Description
	skipInstallationEqual := actual.SkipInstallation == expected.SkipInstallation
	servicesEqual := compareServices(actual.Services, expected.Services, comparer, cli)
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
