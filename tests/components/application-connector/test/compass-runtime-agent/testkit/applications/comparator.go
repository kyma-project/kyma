package applications

import (
	"context"
	"errors"
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate mockery --name=ApplicationGetter
type ApplicationGetter interface {
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.Application, error)
}

func NewComparator(assertions *assert.Assertions, secretComparer Comparator, applicationGetter ApplicationGetter, expectedNamespace, actualNamespace string) (Comparator, error) {
	return &comparator{
		assertions:        assertions,
		secretComparer:    secretComparer,
		applicationGetter: applicationGetter,
		expectedNamespace: expectedNamespace,
		actualNamespace:   actualNamespace,
	}, nil
}

type comparator struct {
	assertions        *assert.Assertions
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

	c.compareSpec(expectedApp, actualApp)
	return nil
}

func (c comparator) compareSpec(expected, actual *v1alpha1.Application) {

	a := c.assertions
	a.Equal(expected.Spec.Description, actual.Spec.Description)
	a.Equal(expected.Spec.SkipInstallation, actual.Spec.SkipInstallation)

	c.compareServices(expected.Spec.Services, actual.Spec.Services)

	a.NotNil(actual.Labels)
	a.Equal(actual.Name, actual.Spec.Labels["connected-app"])

	a.Equal(expected.Spec.Tenant, actual.Spec.Tenant)
	a.Equal(expected.Spec.Group, actual.Spec.Group)

	a.Equal(expected.Spec.Tags, actual.Spec.Tags)
	a.Equal(expected.Spec.DisplayName, actual.Spec.DisplayName)
	a.Equal(expected.Spec.ProviderDisplayName, actual.Spec.ProviderDisplayName)
	a.Equal(expected.Spec.LongDescription, actual.Spec.LongDescription)
	a.Equal(expected.Spec.SkipVerify, actual.Spec.SkipVerify)
}

func (c comparator) compareServices(expected, actual []v1alpha1.Service) {

	a := c.assertions
	a.Equal(len(expected), len(actual))

	for i := 0; i < len(actual); i++ {
		a.Equal(expected[i].Name, actual[i].Name)
		a.Equal(expected[i].Identifier, actual[i].Identifier)
		a.Equal(expected[i].DisplayName, actual[i].DisplayName)
		a.Equal(expected[i].Description, actual[i].Description)

		c.compareEntries(expected[i].Entries, actual[i].Entries)

		a.Equal(expected[i].AuthCreateParameterSchema, actual[i].AuthCreateParameterSchema)
	}
}

func (c comparator) compareEntries(expected, actual []v1alpha1.Entry) {

	a := c.assertions
	a.Equal(len(expected), len(actual))

	for i := 0; i < len(actual); i++ {
		a.Equal(expected[i].Type, actual[i].Type)
		a.Equal(expected[i].TargetUrl, actual[i].TargetUrl)
		a.Equal(expected[i].SpecificationUrl, actual[i].SpecificationUrl)
		a.Equal(expected[i].ApiType, actual[i].ApiType)

		c.compareCredentials(expected[i].Credentials, actual[i].Credentials)

		a.Equal(expected[i].RequestParametersSecretName, actual[i].RequestParametersSecretName)
		a.Equal(expected[i].Name, actual[i].Name)
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
