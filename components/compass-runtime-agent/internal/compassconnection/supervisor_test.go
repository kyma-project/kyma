package compassconnection

import (
	"testing"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/apperrors"

	kymamodel "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/model"

	syncMocks "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/mocks"

	mocks2 "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/compassconnection/mocks"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/certificates"

	certMocks "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/certificates/mocks"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/compass"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/pkg/apis/compass/v1alpha1"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/compass/mocks"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/pkg/client/clientset/versioned/fake"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

const (
	directorURL = "https://director.com/graphql"
)

func TestCrSupervisor_InitializeCompassConnectionCR(t *testing.T) {

	for _, testCase := range []struct {
		description        string
		existingConnection *v1alpha1.CompassConnection
		expectedState      v1alpha1.ConnectionState
	}{
		{
			description:        "CR does not exist",
			existingConnection: nil,
			expectedState:      v1alpha1.Connected,
		},
		{
			description: "not connected CR exists",
			existingConnection: &v1alpha1.CompassConnection{
				ObjectMeta: v1.ObjectMeta{Name: DefaultCompassConnectionName},
				Status:     v1alpha1.CompassConnectionStatus{State: v1alpha1.ConnectionFailed},
			},
			expectedState: v1alpha1.Connected,
		},
		{
			description: "connected CR exists",
			existingConnection: &v1alpha1.CompassConnection{
				ObjectMeta: v1.ObjectMeta{Name: DefaultCompassConnectionName},
				Status:     v1alpha1.CompassConnectionStatus{State: v1alpha1.Connected},
			},
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

			establishedConnection := compass.EstablishedConnection{
				Credentials: certificates.Credentials{},
				DirectorURL: directorURL,
			}

			connector := &mocks.Connector{}
			connector.On("EstablishConnection").Return(establishedConnection, nil)

			certManager := &certMocks.Manager{}
			certManager.On("PreserveCredentials", establishedConnection.Credentials).Return(nil)

			supervisor := NewSupervisor(connector, fakeCRDClient, certManager, nil, nil)

			// when
			connectionCR, err := supervisor.InitializeCompassConnection()

			// then
			require.NoError(t, err)
			createdConnection, err := fakeCRDClient.Get(DefaultCompassConnectionName, v1.GetOptions{})
			require.NoError(t, err)

			assert.Equal(t, testCase.expectedState, createdConnection.Status.State)
			assert.Equal(t, connectionCR, createdConnection)
		})

	}

	t.Run("should set Connection error status when failed to preserve certificate", func(t *testing.T) {
		// given
		fakeCRDClient := fake.NewSimpleClientset().Compass().CompassConnections()

		establishedConnection := compass.EstablishedConnection{
			Credentials: certificates.Credentials{},
			DirectorURL: directorURL,
		}

		connector := &mocks.Connector{}
		connector.On("EstablishConnection").Return(establishedConnection, nil)
		certManager := &certMocks.Manager{}
		certManager.On("PreserveCredentials", establishedConnection.Credentials).Return(errors.New("Error"))

		supervisor := NewSupervisor(connector, fakeCRDClient, certManager, nil, nil)

		// when
		_, err := supervisor.InitializeCompassConnection()

		// then
		require.NoError(t, err)
		assertConnectionFailed(t, fakeCRDClient, DefaultCompassConnectionName)
	})

	t.Run("should set Connection error status when failed to connect to compass", func(t *testing.T) {
		// given
		fakeCRDClient := fake.NewSimpleClientset().Compass().CompassConnections()

		connector := &mocks.Connector{}
		connector.On("EstablishConnection").Return(compass.EstablishedConnection{}, errors.New("error"))

		supervisor := NewSupervisor(connector, fakeCRDClient, nil, nil, nil)

		// when
		_, err := supervisor.InitializeCompassConnection()

		// then
		require.NoError(t, err)
		assertConnectionFailed(t, fakeCRDClient, DefaultCompassConnectionName)
	})

}

func assertConnectionFailed(t *testing.T, client CRManager, connectionName string) {
	createdConnection, err := client.Get(connectionName, v1.GetOptions{})
	require.NoError(t, err)

	assert.Equal(t, v1alpha1.ConnectionFailed, createdConnection.Status.State)
	assert.NotEmpty(t, createdConnection.Status.ConnectionStatus.Error)
}

func TestCrSupervisor_InitializeCompassConnectionCR_Error(t *testing.T) {

	t.Run("should return error when failed to create CR", func(t *testing.T) {
		// given
		mockCRDClient := &mocks2.CRManager{}
		mockCRDClient.On("Get", DefaultCompassConnectionName, v1.GetOptions{}).
			Return(nil, k8serrors.NewNotFound(schema.GroupResource{}, "error"))
		mockCRDClient.On("Create", mock.AnythingOfType("*v1alpha1.CompassConnection")).
			Return(nil, errors.New("error"))

		establishedConnection := compass.EstablishedConnection{
			Credentials: certificates.Credentials{},
			DirectorURL: directorURL,
		}

		connector := &mocks.Connector{}
		connector.On("EstablishConnection").Return(establishedConnection, nil)
		certManager := &certMocks.Manager{}
		certManager.On("PreserveCredentials", establishedConnection.Credentials).Return(nil)

		supervisor := NewSupervisor(connector, mockCRDClient, certManager, nil, nil)

		// when
		_, err := supervisor.InitializeCompassConnection()

		// then
		require.Error(t, err)
	})

	t.Run("should return error when failed to get CR", func(t *testing.T) {
		// given
		mockCRDClient := &mocks2.CRManager{}
		mockCRDClient.On("Get", DefaultCompassConnectionName, v1.GetOptions{}).
			Return(nil, errors.New("error"))

		supervisor := NewSupervisor(nil, mockCRDClient, nil, nil, nil)

		// when
		_, err := supervisor.InitializeCompassConnection()

		// then
		require.Error(t, err)
	})
}

func TestCrSupervisor_SynchronizeWithCompass(t *testing.T) {

	existingConnection := &v1alpha1.CompassConnection{
		ObjectMeta: v1.ObjectMeta{
			Name: DefaultCompassConnectionName,
		},
		Spec: v1alpha1.CompassConnectionSpec{
			ManagementInfo: v1alpha1.ManagementInfo{
				DirectorURL: directorURL,
			},
		},
	}

	credentials := certificates.Credentials{}
	applicationsConfig := []kymamodel.Application{}
	syncResults := []kyma.Result{}

	for _, testCase := range []struct {
		description             string
		mockFunc                func(credManager, compassClient, syncService *mock.Mock)
		expectedConnectionState v1alpha1.ConnectionState
		statusCheckFunc         func(t *testing.T, status *v1alpha1.SynchronizationStatus)
	}{
		{
			description: "synchronize and apply config",
			mockFunc: func(credManager, compassClient, syncService *mock.Mock) {
				credManager.On("GetCredentials").Return(credentials, nil)
				compassClient.On("FetchConfiguration", directorURL, credentials).
					Return(applicationsConfig, nil)
				syncService.On("Apply", applicationsConfig).
					Return(syncResults, nil)
			},
			expectedConnectionState: v1alpha1.Synchronized,
			statusCheckFunc: func(t *testing.T, status *v1alpha1.SynchronizationStatus) {
				assert.Empty(t, status.Error)
				assert.NotEmpty(t, status.LastSuccessfulFetch)
				assert.NotEmpty(t, status.LastSuccessfulApplication)
			},
		},
		{
			description: "set error status if failed to apply config",
			mockFunc: func(credManager, compassClient, syncService *mock.Mock) {
				credManager.On("GetCredentials").Return(credentials, nil)
				compassClient.On("FetchConfiguration", directorURL, credentials).
					Return(applicationsConfig, nil)
				syncService.On("Apply", applicationsConfig).
					Return(syncResults, apperrors.Internal("error"))
			},
			expectedConnectionState: v1alpha1.ResourceApplicationFailed,
			statusCheckFunc: func(t *testing.T, status *v1alpha1.SynchronizationStatus) {
				assert.NotEmpty(t, status.Error)
				assert.NotEmpty(t, status.LastSuccessfulFetch)
				assert.Empty(t, status.LastSuccessfulApplication)
			},
		},
		{
			description: "set error status if failed to fetch config",
			mockFunc: func(credManager, compassClient, syncService *mock.Mock) {
				credManager.On("GetCredentials").Return(credentials, nil)
				compassClient.On("FetchConfiguration", directorURL, credentials).
					Return(nil, errors.New("error"))
			},
			expectedConnectionState: v1alpha1.SynchronizationFailed,
			statusCheckFunc: func(t *testing.T, status *v1alpha1.SynchronizationStatus) {
				assert.NotEmpty(t, status.Error)
				assert.Empty(t, status.LastSuccessfulFetch)
				assert.Empty(t, status.LastSuccessfulApplication)
			},
		},
		{
			description: "set error status if failed to get credentials",
			mockFunc: func(credManager, compassClient, syncService *mock.Mock) {
				credManager.On("GetCredentials").
					Return(certificates.Credentials{}, errors.New("error"))
			},
			expectedConnectionState: v1alpha1.SynchronizationFailed,
			statusCheckFunc: func(t *testing.T, status *v1alpha1.SynchronizationStatus) {
				assert.NotEmpty(t, status.Error)
				assert.Empty(t, status.LastSuccessfulFetch)
				assert.Empty(t, status.LastSuccessfulApplication)
			},
		},
	} {
		t.Run("should "+testCase.description, func(t *testing.T) {
			// given
			fakeCRDClient := fake.NewSimpleClientset(existingConnection).Compass().CompassConnections()

			credManager := &certMocks.Manager{}
			compassClient := &mocks.ConfigClient{}
			syncService := &syncMocks.Service{}

			testCase.mockFunc(&credManager.Mock, &compassClient.Mock, &syncService.Mock)

			supervisor := NewSupervisor(nil, fakeCRDClient, credManager, compassClient, syncService)

			// when
			updatedConnection, err := supervisor.SynchronizeWithCompass(existingConnection)

			// then
			require.NoError(t, err)

			createdConnection, err := fakeCRDClient.Get(DefaultCompassConnectionName, v1.GetOptions{})
			require.NoError(t, err)

			assert.Equal(t, testCase.expectedConnectionState, createdConnection.Status.State)
			assert.NotEmpty(t, createdConnection.Status.SynchronizationStatus.LastAttempt)
			testCase.statusCheckFunc(t, createdConnection.Status.SynchronizationStatus)

			assert.Equal(t, createdConnection, updatedConnection)
		})
	}

}
