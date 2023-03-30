package applications

import (
	"errors"
	"testing"

	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/apis/applicationconnector/v1alpha1"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/kyma/tests/components/application-connector/test/compass-runtime-agent/testkit/applications/mocks"
)

func TestApplicationCrdCompare(t *testing.T) {

	t.Run("should compare applications", func(t *testing.T) {
		secretComparatorMock := &mocks.Comparator{}
		applicationGetterMock := &mocks.ApplicationGetter{}
		actualApp := getTestApp("actual", "actualNamespace", "actualSecret")
		expectedApp := getTestApp("expected", "expectedNamespace", "expectedSecret")

		secretComparatorMock.On("Compare", mock.Anything, "expectedSecret", "actualSecret").Return(nil)
		applicationGetterMock.On("Get", mock.Anything, "actual", v1.GetOptions{}).Return(actualApp, nil).Once()
		applicationGetterMock.On("Get", mock.Anything, "expected", v1.GetOptions{}).Return(expectedApp, nil).Once()

		//when
		applicationComparator, err := NewComparator(secretComparatorMock, applicationGetterMock, "expectedNamespace", "actualNamespace")
		err = applicationComparator.Compare(t, "expected", "actual")

		//then
		require.NoError(t, err)
		secretComparatorMock.AssertExpectations(t)
		applicationGetterMock.AssertExpectations(t)
	})

	t.Run("should return error when expected or actual application name is empty", func(t *testing.T) {
		//given
		secretComparatorMock := &mocks.Comparator{}
		applicationGetterMock := &mocks.ApplicationGetter{}

		{
			//when
			applicationComparator, err := NewComparator(secretComparatorMock, applicationGetterMock, "expected", "actual")
			err = applicationComparator.Compare(t, "expected", "")

			//then
			require.Error(t, err)
		}

		{
			//when
			applicationComparator, err := NewComparator(secretComparatorMock, applicationGetterMock, "expected", "actual")
			err = applicationComparator.Compare(t, "", "actual")

			//then
			require.Error(t, err)
		}

	})

	t.Run("should return error when failed to get actual application", func(t *testing.T) {
		//given
		secretComparatorMock := &mocks.Comparator{}
		applicationGetterMock := &mocks.ApplicationGetter{}
		actualApp := v1alpha1.Application{}

		applicationGetterMock.On("Get", mock.Anything, "actual", v1.GetOptions{}).Return(&actualApp, errors.New("failed to get actual app")).Once()

		//when
		applicationComparator, err := NewComparator(secretComparatorMock, applicationGetterMock, "expected", "actual")
		err = applicationComparator.Compare(t, "expected", "actual")

		//then
		require.Error(t, err)
		secretComparatorMock.AssertExpectations(t)
		applicationGetterMock.AssertExpectations(t)
	})

	t.Run("should return error when failed to get expected application", func(t *testing.T) {
		//given
		secretComparatorMock := &mocks.Comparator{}
		applicationGetterMock := &mocks.ApplicationGetter{}
		expectedApp := v1alpha1.Application{}
		actualApp := v1alpha1.Application{}

		applicationGetterMock.On("Get", mock.Anything, "actual", v1.GetOptions{}).Return(&actualApp, nil).Once()
		applicationGetterMock.On("Get", mock.Anything, "expected", v1.GetOptions{}).Return(&expectedApp, errors.New("failed to get expected app")).Once()

		//when
		applicationComparator, err := NewComparator(secretComparatorMock, applicationGetterMock, "expected", "actual")
		err = applicationComparator.Compare(t, "expected", "actual")

		//then
		require.Error(t, err)
		secretComparatorMock.AssertExpectations(t)
		applicationGetterMock.AssertExpectations(t)
	})
}

func getTestApp(name, namespace, secretName string) *v1alpha1.Application {
	//given
	services := make([]v1alpha1.Service, 0, 0)
	entries := make([]v1alpha1.Entry, 0, 0)

	credentials := v1alpha1.Credentials{
		Type:              "OAuth",
		SecretName:        secretName,
		AuthenticationUrl: "authURL",
		CSRFInfo:          &v1alpha1.CSRFInfo{TokenEndpointURL: "csrfTokenURL"},
	}

	entries = append(entries, v1alpha1.Entry{
		Type:                        "api",
		TargetUrl:                   "targetURL",
		SpecificationUrl:            "specURL",
		ApiType:                     "v1",
		Credentials:                 credentials,
		RequestParametersSecretName: "paramSecret",
		Name:                        "test2",
		ID:                          "t2",
		CentralGatewayUrl:           "centralURL",
		AccessLabel:                 "", //ignore for now
		GatewayUrl:                  "",
	})

	entries = append(entries, v1alpha1.Entry{
		Type:                        "api",
		TargetUrl:                   "targetURL",
		SpecificationUrl:            "specURL",
		ApiType:                     "v1",
		Credentials:                 credentials,
		RequestParametersSecretName: "paramSecret",
		Name:                        "test1",
		ID:                          "t1",
		CentralGatewayUrl:           "centralURL",
		AccessLabel:                 "",
		GatewayUrl:                  "",
	})

	services = append(services, v1alpha1.Service{
		ID:                        "serviceTest",
		Identifier:                "st1",
		Name:                      "srvTest1",
		DisplayName:               "srvTest1",
		Description:               "srvTest1",
		Entries:                   entries,
		AuthCreateParameterSchema: nil,
		Labels:                    nil,
		LongDescription:           "",
		ProviderDisplayName:       "",
		Tags:                      nil,
	})

	services = append(services, v1alpha1.Service{
		ID:                        "serviceTest2",
		Identifier:                "st2",
		Name:                      "srvTest2",
		DisplayName:               "srvTest2",
		Description:               "srvTest2",
		Entries:                   entries,
		AuthCreateParameterSchema: nil,
		Labels:                    nil,
		LongDescription:           "",
		ProviderDisplayName:       "",
		Tags:                      nil,
	})

	return &v1alpha1.Application{
		TypeMeta: v1.TypeMeta{},
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.ApplicationSpec{
			Description:      "testapp",
			SkipInstallation: false,
			Services:         services,
			Labels:           map[string]string{"connected-app": name},

			Tenant: "test",
			Group:  "test",
			CompassMetadata: &v1alpha1.CompassMetadata{
				ApplicationID:  "compassID1",
				Authentication: v1alpha1.Authentication{ClientIds: []string{"11", "22"}},
			},
			Tags:                []string{"tag1", "tag2"},
			DisplayName:         "applicationOneDisplay",
			ProviderDisplayName: "applicationOneDisplay",
			LongDescription:     "applicationOne Test",
			SkipVerify:          true,
		},
	}

}
