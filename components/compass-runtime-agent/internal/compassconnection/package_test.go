package compassconnection

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/apperrors"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/certificates"

	"github.com/stretchr/testify/assert"

	certsMocks "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/certificates/mocks"
	directorMocks "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/compass/director/mocks"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/config"
	configMocks "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/config/mocks"
	kymaMocks "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/mocks"
	kymaModel "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/model"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	compassMocks "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/compass/mocks"

	"github.com/stretchr/testify/mock"

	"github.com/kyma-incubator/compass/components/connector/pkg/gqlschema"

	connectorMocks "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/compass/connector/mocks"

	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/pkg/apis/compass/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/require"
)

const (
	compassConnectionName = "compass-connection"
	token                 = "token"
	runtimeId             = "abcd-efgh-ijkl"

	syncPeriod            = 2 * time.Second
	minimalConfigSyncTime = 4 * time.Second
)

var (
	connectorTokenHeaders = map[string][]string{
		ConnectorTokenHeader: {token},
	}
	nilHeaders map[string][]string

	connectorURL            = "https://connector.com"
	directorURL             = "https://director.com"
	certSecuredConnectorURL = "https://cert-connector.com"

	connectorConfigurationResponse = gqlschema.Configuration{
		Token: &gqlschema.Token{Token: token},
		CertificateSigningRequestInfo: &gqlschema.CertificateSigningRequestInfo{
			Subject:      "O=Org,OU=OrgUnit,L=locality,ST=province,C=DE,CN=test",
			KeyAlgorithm: "rsa2048",
		},
		ManagementPlaneInfo: &gqlschema.ManagementPlaneInfo{
			DirectorURL:                    &directorURL,
			CertificateSecuredConnectorURL: &certSecuredConnectorURL,
		},
	}

	connectionConfig = config.ConnectionConfig{
		Token:        token,
		ConnectorURL: connectorURL,
	}

	runtimeConfig = config.RuntimeConfig{RuntimeId: runtimeId}

	kymaModelApps = []kymaModel.Application{{Name: "App-1", ID: "abcd-efgh"}}

	operationResults = []kyma.Result{{ApplicationID: "abcd-efgh", Operation: kyma.Create}}
)

func TestCompassConnectionController(t *testing.T) {

	syncPeriodTime := syncPeriod
	ctrlManager, err := manager.New(cfg, manager.Options{SyncPeriod: &syncPeriodTime})
	require.NoError(t, err)

	// Credentials manager
	credentialsManagerMock := &certsMocks.Manager{}
	credentialsManagerMock.On("GetClientCredentials").Return(credentials.ClientCredentials, nil)
	credentialsManagerMock.On("PreserveCredentials", mock.AnythingOfType("certificates.Credentials")).Run(func(args mock.Arguments) {
		credentials, ok := args[0].(certificates.Credentials)
		assert.True(t, ok)
		assert.NotEmpty(t, credentials)
	}).Return(nil)
	// Config provider
	configProviderMock := configProviderMock()
	// Connector clients
	tokensConnectorClientMock := connectorTokensClientMock()
	certsConnectorClientMock := connectorCertClientMock()
	// Director config client
	configurationClientMock := &directorMocks.ConfigClient{}
	configurationClientMock.On("FetchConfiguration", directorURL, runtimeId).Return(kymaModelApps, nil)
	// Clients provider
	clientsProviderMock := clientsProviderMock(configurationClientMock, tokensConnectorClientMock, certsConnectorClientMock)
	// Sync service
	synchronizationServiceMock := &kymaMocks.Service{}
	synchronizationServiceMock.On("Apply", kymaModelApps).Return(operationResults, nil)

	var baseDependencies = DependencyConfig{
		K8sConfig:         cfg,
		ControllerManager: ctrlManager,

		ClientsProvider:              clientsProviderMock,
		CredentialsManager:           credentialsManagerMock,
		SynchronizationService:       synchronizationServiceMock,
		ConfigProvider:               configProviderMock,
		CertValidityRenewalThreshold: 0.3,
		MinimalCompassSyncTime:       minimalConfigSyncTime,
	}

	supervisor, err := baseDependencies.InitializeController()
	require.NoError(t, err)

	stopChan, _ := StartTestManager(t, ctrlManager)
	defer close(stopChan)
	defer func() {
		compassConnectionCRClient.Delete(compassConnectionName, &v1.DeleteOptions{})
	}()

	connection, err := supervisor.InitializeCompassConnection()
	require.NoError(t, err)
	assert.NotEmpty(t, connection)

	t.Run("Compass Connection should be synchronized after few seconds", func(t *testing.T) {
		// when
		waitForResynchronization()

		// then
		assertCompassConnectionState(t, v1alpha1.Synchronized)

		mock.AssertExpectationsForObjects(t,
			tokensConnectorClientMock,
			configurationClientMock,
			synchronizationServiceMock,
			clientsProviderMock,
			configProviderMock,
			credentialsManagerMock)
		certsConnectorClientMock.AssertCalled(t, "Configuration", nilHeaders)
		certsConnectorClientMock.AssertNotCalled(t, "SignCSR", mock.AnythingOfType("string"), nilHeaders)
	})

	t.Run("Compass Connection should be reinitialized if deleted", func(t *testing.T) {
		// given
		err := compassConnectionCRClient.Delete(compassConnectionName, &v1.DeleteOptions{})
		require.NoError(t, err)

		// when
		waitForResynchronization()

		// then
		assertCompassConnectionState(t, v1alpha1.Synchronized)

		mock.AssertExpectationsForObjects(t,
			tokensConnectorClientMock,
			configurationClientMock,
			synchronizationServiceMock,
			clientsProviderMock,
			configProviderMock,
			credentialsManagerMock)
		certsConnectorClientMock.AssertCalled(t, "Configuration", nilHeaders)
		certsConnectorClientMock.AssertNotCalled(t, "SignCSR", mock.AnythingOfType("string"), nilHeaders)
	})

	t.Run("should not reinitialized connection if connection is in Synchronized state", func(t *testing.T) {
		// when
		connection, err := supervisor.InitializeCompassConnection()

		// then
		require.NoError(t, err)
		assert.Equal(t, v1alpha1.Synchronized, connection.Status.State)
	})

	t.Run("Should renew certificate if RefreshCredentialsNow set to true", func(t *testing.T) {
		// given
		connectedConnection, err := compassConnectionCRClient.Get(compassConnectionName, v1.GetOptions{})
		require.NoError(t, err)

		connectedConnection.Spec.RefreshCredentialsNow = true

		// when
		connectedConnection, err = compassConnectionCRClient.Update(connectedConnection)
		require.NoError(t, err)
		waitForResynchronization()

		// then
		assertCompassConnectionState(t, v1alpha1.Synchronized)

		certsConnectorClientMock.AssertCalled(t, "SignCSR", mock.AnythingOfType("string"), nilHeaders)
	})

	t.Run("Compass Connection should be in ResourceApplicationFailed state if failed to apply resources", func(t *testing.T) {
		// given
		synchronizationServiceMock.ExpectedCalls = nil
		synchronizationServiceMock.On("Apply", kymaModelApps).Return(nil, apperrors.Internal("error"))

		// when
		waitForResynchronization()

		// then
		assertCompassConnectionState(t, v1alpha1.ResourceApplicationFailed)
	})

	t.Run("Compass Connection should be in SynchronizationFailed state if failed to fetch configuration from Director", func(t *testing.T) {
		// given
		configurationClientMock.ExpectedCalls = nil
		configurationClientMock.On("FetchConfiguration", directorURL, runtimeId).Return(nil, errors.New("error"))

		// when
		waitForResynchronization()

		// then
		assertCompassConnectionState(t, v1alpha1.SynchronizationFailed)
	})

	t.Run("Compass Connection should be in SynchronizationFailed state if failed create Director config client", func(t *testing.T) {
		// given
		clientsProviderMock.ExpectedCalls = nil
		clientsProviderMock.On("GetConnectorCertSecuredClient", credentials.ClientCredentials, certSecuredConnectorURL).Return(certsConnectorClientMock, nil)
		clientsProviderMock.On("GetCompassConfigClient", credentials.ClientCredentials, directorURL).Return(nil, errors.New("error"))

		// when
		waitForResynchronization()

		// then
		assertCompassConnectionState(t, v1alpha1.SynchronizationFailed)
	})

	t.Run("Compass Connection should be in SynchronizationFailed state if failed to read runtime configuration", func(t *testing.T) {
		// given
		configProviderMock.ExpectedCalls = nil
		configProviderMock.On("GetConnectionConfig").Return(connectionConfig, nil)
		configProviderMock.On("GetRuntimeConfig").Return(runtimeConfig, errors.New("error"))

		// when
		waitForResynchronization()

		// then
		assertCompassConnectionState(t, v1alpha1.SynchronizationFailed)
	})

	t.Run("Compass Connection should be in ConnectionMaintenanceFailed if failed to access Connector Configuration query", func(t *testing.T) {
		// given
		certsConnectorClientMock.ExpectedCalls = nil
		certsConnectorClientMock.On("Configuration", nilHeaders).Return(gqlschema.Configuration{}, errors.New("error"))

		// when
		waitForResynchronization()

		// then
		assertCompassConnectionState(t, v1alpha1.ConnectionMaintenanceFailed)
	})

	t.Run("Compass Connection should be in ConnectionMaintenanceFailed state if failed create Cert secured client", func(t *testing.T) {
		// given
		clientsProviderMock.ExpectedCalls = nil
		clientsProviderMock.On("GetConnectorCertSecuredClient", credentials.ClientCredentials, certSecuredConnectorURL).Return(nil, errors.New("error"))

		// when
		waitForResynchronization()

		// then
		assertCompassConnectionState(t, v1alpha1.ConnectionMaintenanceFailed)
	})

	t.Run("Compass Connection should be in ConnectionMaintenanceFailed if failed to get client credentials from secret", func(t *testing.T) {
		// given
		credentialsManagerMock.ExpectedCalls = nil
		credentialsManagerMock.On("GetClientCredentials").Return(certificates.ClientCredentials{}, errors.New("error"))

		// when
		waitForResynchronization()

		// then
		assertCompassConnectionState(t, v1alpha1.ConnectionMaintenanceFailed)
	})
}

func TestFailedToInitializeConnection(t *testing.T) {

	syncPeriodTime := syncPeriod
	ctrlManager, err := manager.New(cfg, manager.Options{SyncPeriod: &syncPeriodTime})
	require.NoError(t, err)

	// Connector token client
	connectorTokenClientMock := connectorTokensClientMock()
	// Config provider
	configProviderMock := configProviderMock()
	// Clients provider
	clientsProviderMock := clientsProviderMock(nil, connectorTokenClientMock, nil)

	// Credentials manager
	credentialsManagerMock := &certsMocks.Manager{}

	var baseDependencies = DependencyConfig{
		K8sConfig:         cfg,
		ControllerManager: ctrlManager,

		ClientsProvider:              clientsProviderMock,
		CredentialsManager:           credentialsManagerMock,
		SynchronizationService:       nil,
		ConfigProvider:               configProviderMock,
		CertValidityRenewalThreshold: 0.3,
		MinimalCompassSyncTime:       minimalConfigSyncTime,
	}

	supervisor, err := baseDependencies.InitializeController()
	require.NoError(t, err)

	stopChan, _ := StartTestManager(t, ctrlManager)
	defer close(stopChan)
	defer func() {
		compassConnectionCRClient.Delete(compassConnectionName, &v1.DeleteOptions{})
	}()

	initConnectionIfNotExist := func() {
		_, err := compassConnectionCRClient.Get(compassConnectionName, v1.GetOptions{})
		if err != nil {
			if !k8serrors.IsNotFound(err) {
				t.Fatalf("Failed to initialize Connection: %s", err.Error())
			}

			connection, err := supervisor.InitializeCompassConnection()
			require.NoError(t, err)
			assert.NotEmpty(t, connection)
		}
	}

	for _, test := range []struct {
		description string
		setupFunc   func()
	}{
		{
			description: "failed to preserve credentials",
			setupFunc: func() {
				credentialsManagerMock.On("PreserveCredentials", mock.AnythingOfType("certificates.Credentials")).Return(errors.New("error"))
			},
		},
		{
			description: "failed to sign CSR",
			setupFunc: func() {
				connectorTokenClientMock.ExpectedCalls = nil
				connectorTokenClientMock.On("Configuration", connectorTokenHeaders).Return(connectorConfigurationResponse, nil)
				connectorTokenClientMock.On("SignCSR", mock.AnythingOfType("string"), connectorTokenHeaders).Return(gqlschema.CertificationResult{}, errors.New("error"))
			},
		},
		{
			description: "failed to fetch Configuration",
			setupFunc: func() {
				connectorTokenClientMock.ExpectedCalls = nil
				connectorTokenClientMock.On("Configuration", connectorTokenHeaders).Return(gqlschema.Configuration{}, errors.New("error"))
				connectorTokenClientMock.On("SignCSR", mock.AnythingOfType("string"), connectorTokenHeaders).Return(gqlschema.CertificationResult{}, errors.New("error"))
			},
		},
		{
			description: "failed to get Connector client",
			setupFunc: func() {
				clientsProviderMock.ExpectedCalls = nil
				clientsProviderMock.On("GetConnectorClient", connectorURL).Return(nil, errors.New("error"))
			},
		},
		{
			description: "connector URL is empty",
			setupFunc: func() {
				configProviderMock.ExpectedCalls = nil
				configProviderMock.On("GetConnectionConfig").Return(config.ConnectionConfig{Token: token}, nil)
			},
		},
		{
			description: "failed to get connection config",
			setupFunc: func() {
				configProviderMock.ExpectedCalls = nil
				configProviderMock.On("GetConnectionConfig").Return(config.ConnectionConfig{}, errors.New("error"))
			},
		},
	} {
		t.Run("Compass Connection should be in ConnectionFailed state when "+test.description, func(t *testing.T) {
			// given
			test.setupFunc()
			initConnectionIfNotExist()

			// when
			waitForResynchronization()

			// then
			assertCompassConnectionState(t, v1alpha1.ConnectionFailed)
		})
	}

}

func waitForResynchronization() {
	time.Sleep(minimalConfigSyncTime + 3*time.Second)
}

func assertCompassConnectionState(t *testing.T, expectedState v1alpha1.ConnectionState) {
	connectedConnection, err := compassConnectionCRClient.Get(compassConnectionName, v1.GetOptions{})
	require.NoError(t, err)
	assert.Equal(t, expectedState, connectedConnection.Status.State)
}

func clientsProviderMock(configClient *directorMocks.ConfigClient, connectorTokensClient, connectorCertsClient *connectorMocks.Client) *compassMocks.ClientsProvider {
	clientsProviderMock := &compassMocks.ClientsProvider{}
	clientsProviderMock.On("GetCompassConfigClient", credentials.ClientCredentials, directorURL).Return(configClient, nil)
	clientsProviderMock.On("GetConnectorCertSecuredClient", credentials.ClientCredentials, certSecuredConnectorURL).Return(connectorCertsClient, nil)
	clientsProviderMock.On("GetConnectorClient", connectorURL).Return(connectorTokensClient, nil)

	return clientsProviderMock
}

func connectorCertClientMock() *connectorMocks.Client {
	connectorMock := &connectorMocks.Client{}
	connectorMock.On("Configuration", nilHeaders).Return(connectorConfigurationResponse, nil)
	connectorMock.On("SignCSR", mock.AnythingOfType("string"), nilHeaders).Return(connectorCertResponse, nil)

	return connectorMock
}

func connectorTokensClientMock() *connectorMocks.Client {
	connectorMock := &connectorMocks.Client{}
	connectorMock.On("Configuration", connectorTokenHeaders).Return(connectorConfigurationResponse, nil)
	connectorMock.On("SignCSR", mock.AnythingOfType("string"), connectorTokenHeaders).Return(connectorCertResponse, nil)

	return connectorMock
}

func configProviderMock() *configMocks.Provider {
	providerMock := &configMocks.Provider{}
	providerMock.On("GetConnectionConfig").Return(connectionConfig, nil)
	providerMock.On("GetRuntimeConfig").Return(runtimeConfig, nil)

	return providerMock
}

// StartTestManager
func StartTestManager(t *testing.T, mgr manager.Manager) (chan struct{}, *sync.WaitGroup) {
	stop := make(chan struct{})
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := mgr.Start(stop)
		require.NoError(t, err)
	}()
	return stop, wg
}
