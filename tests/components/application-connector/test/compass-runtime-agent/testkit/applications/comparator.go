package applications

import (
	"context"
	"errors"
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate mockery --name=ApplicationGetter
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

func (c comparator) Compare(expected, actual string) error {

	if actual == "" || expected == "" {
		return errors.New("empty actual or expected application name")
	}

	actualApp, err := c.applicationGetter.Get(context.Background(), actual, v1.GetOptions{})
	if err != nil {
		return err
	}

	expectedApp, err := c.applicationGetter.Get(context.Background(), expected, v1.GetOptions{})
	if err != nil {
		return err
	}

	c.assertions.Equal(expectedApp.Name, actualApp.Name)
	c.assertions.Equal(expectedApp.Namespace, actualApp.Namespace)
	c.compareSpec(expectedApp.Spec, actualApp.Spec)
	return nil
}

func (c comparator) compareSpec(expected, actual v1alpha1.ApplicationSpec) {

	a := c.assertions
	a.Equal(expected.Description, actual.Description)
	a.Equal(expected.SkipInstallation, actual.SkipInstallation)

	c.compareServices(expected.Services, actual.Services)

	a.Equal(expected.Labels, actual.Labels)
	a.Equal(expected.Tenant, actual.Tenant)
	a.Equal(expected.Group, actual.Group)

	a.Equal(expected.CompassMetadata, actual.CompassMetadata)

	a.Equal(expected.Tags, actual.Tags)
	a.Equal(expected.DisplayName, actual.DisplayName)
	a.Equal(expected.ProviderDisplayName, actual.ProviderDisplayName)
	a.Equal(expected.LongDescription, actual.LongDescription)
	a.Equal(expected.SkipVerify, actual.SkipVerify)
}

func (c comparator) compareServices(expected, actual []v1alpha1.Service) {

	a := c.assertions
	a.Equal(len(expected), len(actual))

	for i := 0; i < len(actual); i++ {
		a.Equal(expected[i].Name, actual[i].Name)
		a.Equal(expected[i].ID, actual[i].ID)
		a.Equal(expected[i].Identifier, actual[i].Identifier)
		a.Equal(expected[i].DisplayName, actual[i].DisplayName)
		a.Equal(expected[i].Description, actual[i].Description)

		c.compareEntries(expected[i].Entries, actual[i].Entries)

		a.Equal(expected[i].AuthCreateParameterSchema, actual[i].AuthCreateParameterSchema)
	}
}

func (c comparator) compareEntries(expected, actual []v1alpha1.Entry) {

	a := c.assertions
	a.Equal(len(actual), len(expected))

	for i := 0; i < len(actual); i++ {
		a.Equal(expected[i].Type, actual[i].Type)
		a.Equal(expected[i].TargetUrl, actual[i].TargetUrl)
		a.Equal(expected[i].SpecificationUrl, actual[i].SpecificationUrl)
		a.Equal(expected[i].ApiType, actual[i].ApiType)

		c.compareCredentials(expected[i].Credentials, actual[i].Credentials)

		a.Equal(expected[i].RequestParametersSecretName, actual[i].RequestParametersSecretName)
		a.Equal(expected[i].Name, actual[i].Name)
		a.Equal(expected[i].ID, actual[i].ID)
		a.Equal(expected[i].CentralGatewayUrl, actual[i].CentralGatewayUrl)
	}
}

func (c comparator) compareCredentials(expected, actual v1alpha1.Credentials) {

	a := c.assertions

	a.Equal(expected.Type, actual.Type)

	err := c.secretComparer.Compare(expected.SecretName, actual.SecretName)
	a.NoError(err)

	a.Equal(expected.AuthenticationUrl, actual.AuthenticationUrl)

	a.Equal(expected.CSRFInfo, actual.CSRFInfo)
}
