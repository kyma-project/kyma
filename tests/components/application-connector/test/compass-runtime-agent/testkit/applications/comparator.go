package applications

import (
	"context"
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ApplicationGetter interface {
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.Application, error)
}

func NewComparator(assertions *require.Assertions, secretComparer Comparator, applicationGetter ApplicationGetter, expectedNamespace, actualNamespace string) (Comparator, error) {
	return &comparator{
		assertions:        assertions,
		secretComparer:    secretComparer,
		applicationGetter: applicationGetter,
		expectedNamespace: expectedNamespace,
		actualNamespace:   actualNamespace,
	}, nil
}

type comparator struct {
	assertions        *require.Assertions
	secretComparer    Comparator
	applicationGetter ApplicationGetter
	expectedNamespace string
	actualNamespace   string
}

func (c comparator) Compare(actual, expected string) error {

	actualApp, err := c.applicationGetter.Get(context.Background(), actual, v1.GetOptions{})
	if err != nil {
		return err
	}

	expectedApp, err := c.applicationGetter.Get(context.Background(), expected, v1.GetOptions{})
	if err != nil {
		return err
	}

	c.assertions.Equal(actualApp.Name, expectedApp.Name)
	c.assertions.Equal(actualApp.Namespace, expectedApp.Namespace)
	c.compareSpec(actualApp.Spec, expectedApp.Spec)
	return nil
}

func (c comparator) compareSpec(actual, expected v1alpha1.ApplicationSpec) {

	a := c.assertions
	a.Equal(expected.Description, actual.Description)
	a.Equal(expected.SkipInstallation, actual.SkipInstallation)

	c.compareServices(actual.Services, expected.Services)

	a.Equal(expected.Labels, actual.Labels)
	a.Equal(expected.Tenant, actual.Tenant)
	a.Equal(expected.Group, actual.Group)

	c.compareCompassMetadata(actual.CompassMetadata, expected.CompassMetadata)

	a.Equal(expected.Tags, actual.Tags)
	a.Equal(expected.DisplayName, actual.DisplayName)
	a.Equal(expected.ProviderDisplayName, actual.ProviderDisplayName)
	a.Equal(expected.LongDescription, actual.LongDescription)
	a.Equal(expected.SkipVerify, actual.SkipVerify)
}

//func compareLabels(actual, expected map[string]string) bool {
//	if len(actual) != len(expected) {
//		return false
//	}
//
//	return reflect.DeepEqual(actual, expected)
//}

func (c comparator) compareCompassMetadata(actual, expected *v1alpha1.CompassMetadata) {

	a := c.assertions
	a.Equal(actual.ApplicationID, expected.ApplicationID)

	c.compareAuthentication(actual.Authentication, expected.Authentication)
}

func (c comparator) compareAuthentication(actual, expected v1alpha1.Authentication) {

	a := c.assertions
	a.Equal(len(actual.ClientIds), len(expected.ClientIds))
	a.Equal(actual.ClientIds, expected.ClientIds) //deep equal
}

func (c comparator) compareServices(actual, expected []v1alpha1.Service) {

	a := c.assertions
	a.Equal(len(expected), len(actual))

	for i := 0; i < len(actual); i++ {
		a.Equal(expected[i].Name, actual[i].Name)
		a.Equal(expected[i].ID, actual[i].ID)
		a.Equal(expected[i].Identifier, actual[i].Identifier)
		a.Equal(expected[i].DisplayName, actual[i].DisplayName)
		a.Equal(expected[i].Description, actual[i].Description)
		a.Equal(expected[i].Entries, actual[i].Entries)
		a.Equal(expected[i].AuthCreateParameterSchema, actual[i].AuthCreateParameterSchema)
	}
}

func (c comparator) compareEntries(actual, expected []v1alpha1.Entry) {

	a := c.assertions
	a.Equal(len(actual), len(expected))

	for i := 0; i < len(actual); i++ {
		a.Equal(actual[i].Type, actual[i].Type)
		a.Equal(actual[i].TargetUrl, actual[i].TargetUrl)
		a.Equal(actual[i].SpecificationUrl, actual[i].SpecificationUrl)
		a.Equal(actual[i].ApiType, actual[i].ApiType)

		c.compareCredentials(actual[i].Credentials, actual[i].Credentials)

		a.Equal(actual[i].RequestParametersSecretName, actual[i].RequestParametersSecretName)
		a.Equal(actual[i].Name, actual[i].Name)
		a.Equal(actual[i].ID, actual[i].ID)
		a.Equal(actual[i].CentralGatewayUrl, actual[i].CentralGatewayUrl)
	}

}

func (c comparator) compareCredentials(actual, expected v1alpha1.Credentials) {

	a := c.assertions

	if (actual == v1alpha1.Credentials{} && expected == v1alpha1.Credentials{}) {
		return
	}

	a.Equal(actual.Type, expected.Type)

	err := c.secretComparer.Compare(actual.SecretName, "oauth-test-expected")
	a.NoError(err)

	a.Equal(actual.SecretName, expected.SecretName)
	a.Equal(actual.AuthenticationUrl, expected.AuthenticationUrl)

	c.compareCSRF(actual.CSRFInfo, expected.CSRFInfo)

}

func (c comparator) compareCSRF(actual, expected *v1alpha1.CSRFInfo) {
	a := c.assertions
	a.NotNil(actual)
	a.NotNil(expected)

	a.Equal(actual.TokenEndpointURL, expected.TokenEndpointURL)
}
