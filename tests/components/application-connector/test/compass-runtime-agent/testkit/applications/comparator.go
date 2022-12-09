package applications

import (
	"context"
	"errors"
	"testing"

	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate mockery --name=ApplicationGetter
type ApplicationGetter interface {
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.Application, error)
}

func NewComparator(secretComparer Comparator, applicationGetter ApplicationGetter, expectedNamespace, actualNamespace string) (Comparator, error) {
	return &comparator{
		secretComparer:    secretComparer,
		applicationGetter: applicationGetter,
		expectedNamespace: expectedNamespace,
		actualNamespace:   actualNamespace,
	}, nil
}

type comparator struct {
	secretComparer    Comparator
	applicationGetter ApplicationGetter
	expectedNamespace string
	actualNamespace   string
}

func (c comparator) Compare(t *testing.T, expected, actual string) error {
	t.Helper()

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

	c.compareSpec(t, expectedApp, actualApp)
	return nil
}

func (c comparator) compareSpec(t *testing.T, expected, actual *v1alpha1.Application) {
	t.Helper()
	a := assert.New(t)

	a.Equal(expected.Spec.Description, actual.Spec.Description, "Description is incorrect")
	a.Equal(expected.Spec.SkipInstallation, actual.Spec.SkipInstallation, "SkipInstallation is incorrect")

	c.compareServices(t, expected.Spec.Services, actual.Spec.Services)

	a.NotNil(actual.Spec.Labels)
	a.Equal(actual.Name, actual.Spec.Labels["connected-app"])

	a.Equal(expected.Spec.Tenant, actual.Spec.Tenant, "Tenant is incorrect")
	a.Equal(expected.Spec.Group, actual.Spec.Group, "Group is incorrect")

	a.Equal(expected.Spec.Tags, actual.Spec.Tags, "Tags is incorrect")
	a.Equal(expected.Spec.DisplayName, actual.Spec.DisplayName, "DisplayName is incorrect")
	a.Equal(expected.Spec.ProviderDisplayName, actual.Spec.ProviderDisplayName, "ProviderDisplayName is incorrect")
	a.Equal(expected.Spec.LongDescription, actual.Spec.LongDescription, "LongDescription is incorrect")
	a.Equal(expected.Spec.SkipVerify, actual.Spec.SkipVerify, "SkipVerify is incorrect")
}

func (c comparator) compareServices(t *testing.T, expected, actual []v1alpha1.Service) {
	t.Helper()
	a := assert.New(t)

	a.Equal(len(expected), len(actual))

	for i := 0; i < len(actual); i++ {
		a.Equal(expected[i].Name, actual[i].Name)
		a.Equal(expected[i].Identifier, actual[i].Identifier)
		a.Equal(expected[i].DisplayName, actual[i].DisplayName)
		a.Equal(expected[i].Description, actual[i].Description)

		c.compareEntries(t, expected[i].Entries, actual[i].Entries)

		a.Equal(expected[i].AuthCreateParameterSchema, actual[i].AuthCreateParameterSchema)
	}
}

func (c comparator) compareEntries(t *testing.T, expected, actual []v1alpha1.Entry) {
	t.Helper()
	a := assert.New(t)

	a.Equal(len(expected), len(actual))

	for i := 0; i < len(actual); i++ {
		a.Equal(expected[i].Type, actual[i].Type)
		a.Equal(expected[i].TargetUrl, actual[i].TargetUrl)
		a.Equal(expected[i].SpecificationUrl, actual[i].SpecificationUrl)
		a.Equal(expected[i].ApiType, actual[i].ApiType)

		c.compareCredentials(t, expected[i].Credentials, actual[i].Credentials)

		a.Equal(expected[i].RequestParametersSecretName, actual[i].RequestParametersSecretName)
		a.Equal(expected[i].Name, actual[i].Name)
		a.Equal(expected[i].CentralGatewayUrl, actual[i].CentralGatewayUrl)
	}
}

func (c comparator) compareCredentials(t *testing.T, expected, actual v1alpha1.Credentials) {
	t.Helper()
	a := assert.New(t)

	a.Equal(expected.Type, actual.Type)

	err := c.secretComparer.Compare(t, expected.SecretName, actual.SecretName)
	a.NoError(err)

	a.Equal(expected.AuthenticationUrl, actual.AuthenticationUrl)

	a.Equal(expected.CSRFInfo, actual.CSRFInfo)
}
