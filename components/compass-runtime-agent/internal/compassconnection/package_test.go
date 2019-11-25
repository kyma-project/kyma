package compassconnection

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"kyma-project.io/compass-runtime-agent/internal/compass/director"

	"kyma-project.io/compass-runtime-agent/internal/kyma"

	"kyma-project.io/compass-runtime-agent/internal/apperrors"

	"kyma-project.io/compass-runtime-agent/internal/certificates"

	"github.com/stretchr/testify/assert"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	certsMocks "kyma-project.io/compass-runtime-agent/internal/certificates/mocks"
	directorMocks "kyma-project.io/compass-runtime-agent/internal/compass/director/mocks"
	"kyma-project.io/compass-runtime-agent/internal/config"
	configMocks "kyma-project.io/compass-runtime-agent/internal/config/mocks"
	kymaMocks "kyma-project.io/compass-runtime-agent/internal/kyma/mocks"
	kymaModel "kyma-project.io/compass-runtime-agent/internal/kyma/model"

	compassMocks "kyma-project.io/compass-runtime-agent/internal/compass/mocks"

	"github.com/stretchr/testify/mock"

	gqlschema "github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"

	connectorMocks "kyma-project.io/compass-runtime-agent/internal/compass/connector/mocks"

	"sigs.k8s.io/controller-runtime/pkg/manager"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kyma-project.io/compass-runtime-agent/pkg/apis/compass/v1alpha1"

	"github.com/stretchr/testify/require"
)

const (
	compassConnectionName = "compass-connection"
	token                 = "token"
	runtimeId             = "abcd-efgh-ijkl"

	syncPeriod            = 2 * time.Second
	minimalConfigSyncTime = 4 * time.Second

	checkInterval = 2 * time.Second
	testTimeout   = 20 * time.Second
)

var (
	connectorTokenHeaders = map[string]string{
		ConnectorTokenHeader: token,
	}
	nilHeaders map[string]string

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

	runtimeURLsConfig = director.RuntimeURLsConfig{
		EventsURL:  "https://gateway.kyma.local",
		ConsoleURL: "https://console.kyma.local",
	}

	runtimeLabels = graphql.Labels{
		"events":  runtimeURLsConfig.EventsURL,
		"console": runtimeURLsConfig.ConsoleURL,
	}

	kymaModelApps = []kymaModel.Application{{Name: "App-1", ID: "abcd-efgh"}}

	operationResults = []kyma.Result{{ApplicationName: "App-1", ApplicationID: "abcd-efgh", Operation: kyma.Create}}
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
	configurationClientMock := &directorMocks.DirectorClient{}
	configurationClientMock.On("FetchConfiguration").Return(kymaModelApps, nil)
	configurationClientMock.On("SetURLsLabels", runtimeURLsConfig).Return(runtimeLabels, nil)
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

		RuntimeURLsConfig: runtimeURLsConfig,
	}

	supervisor, err := baseDependencies.InitializeController()
	require.NoError(t, err)

	defer func() {
		err := compassConnectionCRClient.Delete(compassConnectionName, &v1.DeleteOptions{})
		if err != nil {
			t.Logf("error while deleting Compass Connection: %s", err.Error())
		}
	}()
	stopChan, _ := StartTestManager(t, ctrlManager)
	defer close(stopChan)

	connection, err := supervisor.InitializeCompassConnection()
	require.NoError(t, err)
	assert.NotEmpty(t, connection)

	t.Run("Compass Connection should be synchronized after few seconds", func(t *testing.T) {
		// when
		waitForResynchronization()

		// then
		assertCompassConnectionState(t, v1alpha1.Synchronized)
		assertConnectionStatusSet(t)
		assertManagementInfoSetInCR(t)

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

		// then
		err = waitFor(checkInterval, testTimeout, func() bool {
			return isConnectionInState(v1alpha1.Synchronized)
		})
		require.NoError(t, err)
		assertConnectionStatusSet(t)
		assertManagementInfoSetInCR(t)

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

		err = waitFor(checkInterval, testTimeout, func() bool {
			fakeT := &testing.T{}
			called := certsConnectorClientMock.AssertCalled(fakeT, "SignCSR", mock.AnythingOfType("string"), nilHeaders)
			return called
		})

		// then
		require.NoError(t, err)
		require.NoError(t, waitForResourceUpdate(v1alpha1.Synchronized))

		assertCertificateRenewed(t)
		assertManagementInfoSetInCR(t)
		certsConnectorClientMock.AssertCalled(t, "SignCSR", mock.AnythingOfType("string"), nilHeaders)
	})

	t.Run("Compass Connection should be in MetadataUpdateFailed state if failed to set labels on Runtime", func(t *testing.T) {
		// given
		clearMockCalls(&configurationClientMock.Mock)
		configurationClientMock.On("FetchConfiguration").Return(kymaModelApps, nil)
		configurationClientMock.On("SetURLsLabels", runtimeURLsConfig).Return(runtimeLabels, apperrors.Internal("error"))

		// when
		err = waitFor(checkInterval, testTimeout, func() bool {
			return mockFunctionCalled(&configurationClientMock.Mock, "SetURLsLabels", runtimeURLsConfig)
		})

		// then
		require.NoError(t, err)
		require.NoError(t, waitForResourceUpdate(v1alpha1.MetadataUpdateFailed))
		assertManagementInfoSetInCR(t)
	})

	t.Run("Compass Connection should be in ResourceApplicationFailed state if failed to apply resources", func(t *testing.T) {
		// given
		clearMockCalls(&synchronizationServiceMock.Mock)
		synchronizationServiceMock.On("Apply", kymaModelApps).Return(nil, apperrors.Internal("error"))

		// when
		err = waitFor(checkInterval, testTimeout, func() bool {
			return mockFunctionCalled(&synchronizationServiceMock.Mock, "Apply", kymaModelApps)
		})

		// then
		require.NoError(t, err)
		require.NoError(t, waitForResourceUpdate(v1alpha1.ResourceApplicationFailed))
		assertManagementInfoSetInCR(t)
		assertSynchronizationStatusError(t)
	})

	t.Run("Compass Connection should be in SynchronizationFailed state if failed to fetch configuration from Director", func(t *testing.T) {
		// given
		clearMockCalls(&configurationClientMock.Mock)
		configurationClientMock.On("FetchConfiguration").Return(nil, errors.New("error"))

		// when
		err = waitFor(checkInterval, testTimeout, func() bool {
			return mockFunctionCalled(&configurationClientMock.Mock, "FetchConfiguration")
		})

		// then
		require.NoError(t, err)
		require.NoError(t, waitForResourceUpdate(v1alpha1.SynchronizationFailed))
		assertManagementInfoSetInCR(t)
		assertSynchronizationStatusError(t)
	})

	t.Run("Compass Connection should be in SynchronizationFailed state if failed create Director config client", func(t *testing.T) {
		// given
		clientsProviderMock.ExpectedCalls = nil
		clientsProviderMock.Calls = nil
		clientsProviderMock.On("GetConnectorClient", connectorURL).Return(tokensConnectorClientMock, nil)
		clientsProviderMock.On("GetConnectorCertSecuredClient", credentials.ClientCredentials, certSecuredConnectorURL).Return(certsConnectorClientMock, nil)
		clientsProviderMock.On("GetDirectorClient", credentials.ClientCredentials, directorURL, runtimeConfig).Return(nil, errors.New("error"))

		// when
		err = waitFor(checkInterval, testTimeout, func() bool {
			return mockFunctionCalled(&clientsProviderMock.Mock, "GetDirectorClient", credentials.ClientCredentials, directorURL, runtimeConfig)
		})

		// then
		require.NoError(t, err)
		require.NoError(t, waitForResourceUpdate(v1alpha1.SynchronizationFailed))
		assertManagementInfoSetInCR(t)
		assertSynchronizationStatusError(t)
	})

	t.Run("Compass Connection should be in SynchronizationFailed state if failed to read runtime configuration", func(t *testing.T) {
		// given
		configProviderMock.ExpectedCalls = nil
		configProviderMock.Calls = nil
		configProviderMock.On("GetConnectionConfig").Return(connectionConfig, nil)
		configProviderMock.On("GetRuntimeConfig").Return(runtimeConfig, errors.New("error"))

		// when
		err = waitFor(checkInterval, testTimeout, func() bool {
			return mockFunctionCalled(&configProviderMock.Mock, "GetRuntimeConfig")
		})

		// then
		require.NoError(t, err)
		require.NoError(t, waitForResourceUpdate(v1alpha1.SynchronizationFailed))
		assertManagementInfoSetInCR(t)
		assertSynchronizationStatusError(t)
	})

	t.Run("Compass Connection should be in ConnectionMaintenanceFailed if Management Info is empty and no error", func(t *testing.T) {
		// given
		certsConnectorClientMock.ExpectedCalls = nil
		certsConnectorClientMock.On("Configuration", nilHeaders).Return(gqlschema.Configuration{ManagementPlaneInfo: nil}, nil)

		// when
		err = waitFor(checkInterval, testTimeout, func() bool {
			return mockFunctionCalled(&certsConnectorClientMock.Mock, "Configuration", nilHeaders)
		})

		// then
		require.NoError(t, err)
		require.NoError(t, waitForResourceUpdate(v1alpha1.ConnectionMaintenanceFailed))
		assertManagementInfoSetInCR(t)
		assertConnectionStatusSet(t)
	})

	t.Run("Compass Connection should be in ConnectionMaintenanceFailed if failed to access Connector Configuration query", func(t *testing.T) {
		// given
		certsConnectorClientMock.ExpectedCalls = nil
		certsConnectorClientMock.On("Configuration", nilHeaders).Return(gqlschema.Configuration{}, errors.New("error"))

		// when
		err = waitFor(checkInterval, testTimeout, func() bool {
			return mockFunctionCalled(&certsConnectorClientMock.Mock, "Configuration", nilHeaders)
		})

		// then
		require.NoError(t, err)
		require.NoError(t, waitForResourceUpdate(v1alpha1.ConnectionMaintenanceFailed))
		assertConnectionStatusSet(t)
	})

	t.Run("Compass Connection should be in ConnectionMaintenanceFailed state if failed create Cert secured client", func(t *testing.T) {
		// given
		clientsProviderMock.ExpectedCalls = nil
		clientsProviderMock.On("GetConnectorCertSecuredClient", credentials.ClientCredentials, certSecuredConnectorURL).Return(nil, errors.New("error"))

		// when
		err = waitFor(checkInterval, testTimeout, func() bool {
			return mockFunctionCalled(&clientsProviderMock.Mock, "GetConnectorCertSecuredClient", credentials.ClientCredentials, certSecuredConnectorURL)
		})

		// then
		require.NoError(t, err)
		require.NoError(t, waitForResourceUpdate(v1alpha1.ConnectionMaintenanceFailed))
		assertConnectionStatusSet(t)
	})

	t.Run("Compass Connection should be in ConnectionMaintenanceFailed if failed to get client credentials from secret", func(t *testing.T) {
		// given
		credentialsManagerMock.ExpectedCalls = nil
		credentialsManagerMock.On("GetClientCredentials").Return(certificates.ClientCredentials{}, errors.New("error"))

		// when
		err = waitFor(checkInterval, testTimeout, func() bool {
			return mockFunctionCalled(&credentialsManagerMock.Mock, "GetClientCredentials")
		})

		// then
		require.NoError(t, err)
		require.NoError(t, waitForResourceUpdate(v1alpha1.ConnectionMaintenanceFailed))
		assertConnectionStatusSet(t)
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

	defer func() {
		err := compassConnectionCRClient.Delete(compassConnectionName, &v1.DeleteOptions{})
		if err != nil {
			t.Logf("error while deleting Compass Connection: %s", err.Error())
		}
	}()
	stopChan, _ := StartTestManager(t, ctrlManager)
	defer close(stopChan)

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
		description  string
		setupFunc    func()
		waitFunction func() bool
	}{
		{
			description: "failed to preserve credentials",
			setupFunc: func() {
				credentialsManagerMock.Calls = nil
				credentialsManagerMock.On("PreserveCredentials", mock.AnythingOfType("certificates.Credentials")).Return(errors.New("error"))
			},
			waitFunction: func() bool {
				return mockFunctionCalled(&credentialsManagerMock.Mock, "PreserveCredentials", mock.AnythingOfType("certificates.Credentials"))
			},
		},
		{
			description: "failed to sign CSR",
			setupFunc: func() {
				clearMockCalls(&connectorTokenClientMock.Mock)
				connectorTokenClientMock.On("Configuration", connectorTokenHeaders).Return(connectorConfigurationResponse, nil)
				connectorTokenClientMock.On("SignCSR", mock.AnythingOfType("string"), connectorTokenHeaders).Return(gqlschema.CertificationResult{}, errors.New("error"))
			},
			waitFunction: func() bool {
				return mockFunctionCalled(&connectorTokenClientMock.Mock, "SignCSR", mock.AnythingOfType("string"), connectorTokenHeaders)
			},
		},
		{
			description: "failed to fetch Configuration",
			setupFunc: func() {
				clearMockCalls(&connectorTokenClientMock.Mock)
				connectorTokenClientMock.On("Configuration", connectorTokenHeaders).Return(gqlschema.Configuration{}, errors.New("error"))
				connectorTokenClientMock.On("SignCSR", mock.AnythingOfType("string"), connectorTokenHeaders).Return(gqlschema.CertificationResult{}, errors.New("error"))
			},
			waitFunction: func() bool {
				return mockFunctionCalled(&connectorTokenClientMock.Mock, "Configuration", connectorTokenHeaders)
			},
		},
		{
			description: "failed to get Connector client",
			setupFunc: func() {
				clearMockCalls(&clientsProviderMock.Mock)
				clientsProviderMock.On("GetConnectorClient", connectorURL).Return(nil, errors.New("error"))
			},
			waitFunction: func() bool {
				return mockFunctionCalled(&connectorTokenClientMock.Mock, "GetConnectorClient", connectorURL)
			},
		},
		{
			description: "connector URL is empty",
			setupFunc: func() {
				clearMockCalls(&configProviderMock.Mock)
				configProviderMock.On("GetConnectionConfig").Return(config.ConnectionConfig{Token: token}, nil)
			},
			waitFunction: func() bool {

				return mockFunctionCalled(&configProviderMock.Mock, "GetConnectionConfig")

			},
		},
		{
			description: "failed to get connection config",
			setupFunc: func() {
				clearMockCalls(&configProviderMock.Mock)
				configProviderMock.On("GetConnectionConfig").Return(config.ConnectionConfig{}, errors.New("error"))
			},
			waitFunction: func() bool {
				fakeT := &testing.T{}
				called := configProviderMock.AssertCalled(fakeT, "GetConnectionConfig")
				return called
			},
		},
	} {
		t.Run("Compass Connection should be in ConnectionFailed state when "+test.description, func(t *testing.T) {
			// given
			test.setupFunc()
			initConnectionIfNotExist()

			// when
			test.waitFunction()

			// then
			require.NoError(t, waitForResourceUpdate(v1alpha1.ConnectionFailed))
		})
	}

}

func waitFor(interval, timeout time.Duration, isDone func() bool) error {
	done := time.After(timeout)

	for {
		if isDone() {
			return nil
		}

		select {
		case <-done:
			return errors.New("timeout waiting for condition")
		default:
			time.Sleep(interval)
		}
	}
}

func clearMockCalls(mock *mock.Mock) {
	mock.ExpectedCalls = nil
	mock.Calls = nil
}

func mockFunctionCalled(mock *mock.Mock, methodName string, arguments ...interface{}) bool {
	fakeT := &testing.T{}
	return mock.AssertCalled(fakeT, methodName, arguments...)
}

func waitForResynchronization() {
	time.Sleep(minimalConfigSyncTime * 2)
}

func waitForResourceUpdate(expectedState v1alpha1.ConnectionState) error {
	// Initial sleep in case if Connection will change from expected state to different
	time.Sleep(2 * time.Second)

	return waitFor(1*time.Second, 4*time.Second, func() bool {
		return isConnectionInState(expectedState)
	})
}

func isConnectionInState(expectedState v1alpha1.ConnectionState) bool {
	connectedConnection, err := compassConnectionCRClient.Get(compassConnectionName, v1.GetOptions{})
	if err != nil {
		return false
	}

	return connectedConnection.Status.State == expectedState
}

func assertConnectionStatusError(t *testing.T) {
	connectedConnection, err := compassConnectionCRClient.Get(compassConnectionName, v1.GetOptions{})
	require.NoError(t, err)
	t.Logf("Connection status error: %s", connectedConnection.Status.ConnectionStatus.Error)
	assert.NotEmpty(t, connectedConnection.Status.ConnectionStatus.Error)
}

func assertSynchronizationStatusError(t *testing.T) {
	connectedConnection, err := compassConnectionCRClient.Get(compassConnectionName, v1.GetOptions{})
	require.NoError(t, err)
	t.Logf("Synchronization status error: %s", connectedConnection.Status.SynchronizationStatus.Error)
	assert.NotEmpty(t, connectedConnection.Status.SynchronizationStatus.Error)
}

func assertManagementInfoSetInCR(t *testing.T) {
	connectedConnection, err := compassConnectionCRClient.Get(compassConnectionName, v1.GetOptions{})
	require.NoError(t, err)
	assert.Equal(t, directorURL, connectedConnection.Spec.ManagementInfo.DirectorURL)
	assert.Equal(t, certSecuredConnectorURL, connectedConnection.Spec.ManagementInfo.ConnectorURL)
}

func assertCompassConnectionState(t *testing.T, expectedState v1alpha1.ConnectionState) {
	connectedConnection, err := compassConnectionCRClient.Get(compassConnectionName, v1.GetOptions{})
	require.NoError(t, err)
	assert.Equal(t, expectedState, connectedConnection.Status.State)
}

func assertConnectionStatusSet(t *testing.T) {
	connectedConnection, err := compassConnectionCRClient.Get(compassConnectionName, v1.GetOptions{})
	require.NoError(t, err)
	require.NotNil(t, connectedConnection.Status.ConnectionStatus)
	assert.NotEmpty(t, connectedConnection.Status.ConnectionStatus)
	assert.NotEmpty(t, connectedConnection.Status.ConnectionStatus.CertificateStatus)
}

func assertCertificateRenewed(t *testing.T) {
	connectedConnection, err := compassConnectionCRClient.Get(compassConnectionName, v1.GetOptions{})
	require.NoError(t, err)
	require.NotNil(t, connectedConnection.Status.ConnectionStatus)
	assert.NotEmpty(t, connectedConnection.Status.ConnectionStatus.CertificateStatus)
	assert.True(t, connectedConnection.Status.ConnectionStatus.Established.Unix() < connectedConnection.Status.ConnectionStatus.Renewed.Unix())
}

func clientsProviderMock(configClient *directorMocks.DirectorClient, connectorTokensClient, connectorCertsClient *connectorMocks.Client) *compassMocks.ClientsProvider {
	clientsProviderMock := &compassMocks.ClientsProvider{}
	clientsProviderMock.On("GetDirectorClient", credentials.ClientCredentials, directorURL, runtimeConfig).Return(configClient, nil)
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
