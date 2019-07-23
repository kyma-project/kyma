package compassconnection

import (
	"testing"

	"k8s.io/apimachinery/pkg/runtime/schema"

	mocks2 "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/compassconnection/mocks"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/compass"

	"github.com/pkg/errors"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/pkg/apis/compass/v1alpha1"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/mock"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/compass/mocks"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/pkg/client/clientset/versioned/fake"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

func TestCrSupervisor_InitializeCompassConnectionCR(t *testing.T) {

	for _, testCase := range []struct {
		description        string
		existingConnection *v1alpha1.CompassConnection
		connectorMock      compass.Connector
		expectedState      v1alpha1.ConnectionState
	}{
		{
			description:        "CR does not exist and connection was successful",
			existingConnection: nil,
			connectorMock:      workingConnector(),
			expectedState:      v1alpha1.Connected,
		},
		{
			description:        "CR does not exist and connection failed",
			existingConnection: nil,
			connectorMock:      failingConnector(),
			expectedState:      v1alpha1.NotConnected,
		},
		{
			description: "CR exists",
			existingConnection: &v1alpha1.CompassConnection{
				ObjectMeta: v1.ObjectMeta{Name: DefaultCompassConnectionName},
				Status:     v1alpha1.CompassConnectionStatus{ConnectionState: v1alpha1.Connected},
			},
			connectorMock: nil,
			expectedState: v1alpha1.Connected,
		},
	} {
		t.Run("should initialize Compass Connection when "+testCase.description, func(t *testing.T) {
			// given
			fakeCRDClient := fake.NewSimpleClientset().Compass().CompassConnections()
			if testCase.existingConnection != nil {
				_, err := fakeCRDClient.Create(testCase.existingConnection)
				require.NoError(t, err)
			}

			supervisor := NewSupervisor(testCase.connectorMock, fakeCRDClient)

			// when
			err := supervisor.InitializeCompassConnectionCR()

			// then
			require.NoError(t, err)
			createdConnection, err := fakeCRDClient.Get(DefaultCompassConnectionName, v1.GetOptions{})
			require.NoError(t, err)

			assert.Equal(t, testCase.expectedState, createdConnection.Status.ConnectionState)
		})

	}

}

func failingConnector() compass.Connector {
	connector := &mocks.Connector{}
	connector.On("EstablishConnection").Return(errors.New("error"))
	return connector
}

func workingConnector() compass.Connector {
	connector := &mocks.Connector{}
	connector.On("EstablishConnection").Return(nil)
	return connector
}

func TestCrSupervisor_InitializeCompassConnectionCR_Error(t *testing.T) {

	t.Run("should return error when failed to create CR", func(t *testing.T) {
		// given
		mockCRDClient := &mocks2.CRManager{}
		mockCRDClient.On("Get", DefaultCompassConnectionName, v1.GetOptions{}).
			Return(nil, k8serrors.NewNotFound(schema.GroupResource{}, "error"))
		mockCRDClient.On("Create", mock.AnythingOfType("*v1alpha1.CompassConnection")).
			Return(nil, errors.New("error"))

		connector := &mocks.Connector{}
		connector.On("EstablishConnection").Return(nil)

		supervisor := NewSupervisor(connector, mockCRDClient)

		// when
		err := supervisor.InitializeCompassConnectionCR()

		// then
		require.Error(t, err)
	})

	t.Run("should return error when failed to get CR", func(t *testing.T) {
		// given
		mockCRDClient := &mocks2.CRManager{}
		mockCRDClient.On("Get", DefaultCompassConnectionName, v1.GetOptions{}).
			Return(nil, errors.New("error"))

		supervisor := NewSupervisor(nil, mockCRDClient)

		// when
		err := supervisor.InitializeCompassConnectionCR()

		// then
		require.Error(t, err)
	})
}
