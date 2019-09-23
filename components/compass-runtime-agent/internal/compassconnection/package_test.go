package compassconnection

import (
	"errors"
	"sync"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/types"

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
)

func TestCompassConnectionController(t *testing.T) {

	syncPeriodTime := syncPeriod
	ctrlManager, err := manager.New(cfg, manager.Options{SyncPeriod: &syncPeriodTime})
	require.NoError(t, err)

	tokensConnectorClientMock := connectorTokensClientMock(nil, nil)
	certsConnectorClientMock := connectorCertClientMock(nil, nil)
	configurationClientMock := directorClientMock(nil)
	synchronizationServiceMock := synchronizationServiceMock(nil)
	clientsProviderMock := clientsProviderMock(configurationClientMock, tokensConnectorClientMock, certsConnectorClientMock)
	credentialsManagerMock := credentialsManagerMock(t, nil, nil)
	configProviderMock := configProviderMock(nil, nil)

	//// Director config client
	//directorClientMock := &directorMocks.ConfigClient{}
	//directorClientMock.On("FetchConfiguration", directorURL, runtimeId).Return(kymaModelApps, nil)
	//
	//// Connector token client
	//connectorTokenClientMock := &connectorMocks.Client{}
	//connectorTokenClientMock.On("Configuration", connectorTokenHeaders).Return(connectorConfigurationResponse, nil)
	//connectorTokenClientMock.On("SignCSR", mock.AnythingOfType("string"), connectorTokenHeaders).Return(connectorCertResponse, nil)
	//
	//// Clients provider
	//clientsProviderMock := &compassMocks.ClientsProvider{}
	//clientsProviderMock.On("GetCompassConfigClient", credentials.ClientCredentials, directorURL).Return(directorClientMock, nil)
	//clientsProviderMock.On("GetConnectorCertSecuredClient", credentials.ClientCredentials, certSecuredConnectorURL).Return(connectorCertsClient, nil)
	//clientsProviderMock.On("GetConnectorClient", connectorURL).Return(connectorTokenClientMock, nil)

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
	connectorTokenClientMock := &connectorMocks.Client{}
	connectorTokenClientMock.On("Configuration", connectorTokenHeaders).Return(connectorConfigurationResponse, nil)
	connectorTokenClientMock.On("SignCSR", mock.AnythingOfType("string"), connectorTokenHeaders).Return(connectorCertResponse, nil)

	// Config provider
	configProviderMock := &configMocks.Provider{}
	configProviderMock.On("GetConnectionConfig").Return(connectionConfig, nil)

	// Clients provider
	clientsProviderMock := &compassMocks.ClientsProvider{}
	clientsProviderMock.On("GetConnectorClient", connectorURL).Return(connectorTokenClientMock, nil)

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
				connectorTokenClientMock.On("SignCSR", mock.AnythingOfType("string"), connectorTokenHeaders).Return(nil, errors.New("error"))
			},
		},
		{
			description: "failed to fetch Configuration",
			setupFunc: func() {
				connectorTokenClientMock.On("Configuration", connectorTokenHeaders).Return(nil, errors.New("error"))
			},
		},
		{
			description: "failed to get Connector client",
			setupFunc: func() {
				clientsProviderMock.On("GetConnectorClient", connectorURL).Return(connectorTokenClientMock, nil)
			},
		},
		{
			description: "failed to get connection config",
			setupFunc: func() {
				configProviderMock.On("GetConnectionConfig").Return(nil, errors.New("error"))
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

// TODO - cleanup mock functions
func clientsProviderMock(configClient *directorMocks.ConfigClient, connectorTokensClient, connectorCertsClient *connectorMocks.Client) *compassMocks.ClientsProvider {
	clientsProviderMock := &compassMocks.ClientsProvider{}
	clientsProviderMock.On("GetCompassConfigClient", credentials.ClientCredentials, directorURL).Return(configClient, nil)
	clientsProviderMock.On("GetConnectorCertSecuredClient", credentials.ClientCredentials, certSecuredConnectorURL).Return(connectorCertsClient, nil)
	clientsProviderMock.On("GetConnectorClient", connectorURL).Return(connectorTokensClient, nil)

	return clientsProviderMock
}

func directorClientMock(configFetchError error) *directorMocks.ConfigClient {
	directorClientMock := &directorMocks.ConfigClient{}
	directorClientMock.On("FetchConfiguration", directorURL, runtimeId).Return(kymaModelApps, configFetchError)

	return directorClientMock
}

func connectorCertClientMock(configError, signingError error) *connectorMocks.Client {
	connectorMock := &connectorMocks.Client{}
	connectorMock.On("Configuration", nilHeaders).Return(connectorConfigurationResponse, configError)
	connectorMock.On("SignCSR", mock.AnythingOfType("string"), nilHeaders).Return(connectorCertResponse, signingError) // TODO - renewed creds?

	return connectorMock
}

func connectorTokensClientMock(configErr, signingErr error) *connectorMocks.Client {
	connectorMock := &connectorMocks.Client{}
	connectorMock.On("Configuration", connectorTokenHeaders).Return(connectorConfigurationResponse, configErr)
	connectorMock.On("SignCSR", mock.AnythingOfType("string"), connectorTokenHeaders).Return(connectorCertResponse, signingErr)

	return connectorMock
}

func credentialsManagerMock(t *testing.T, getCredsError, preserveError error) *certsMocks.Manager {
	credManagerMock := &certsMocks.Manager{}
	credManagerMock.On("GetClientCredentials").Return(credentials.ClientCredentials, getCredsError)
	credManagerMock.On("PreserveCredentials", mock.AnythingOfType("certificates.Credentials")).Run(func(args mock.Arguments) {
		credentials, ok := args[0].(certificates.Credentials)
		assert.True(t, ok)
		assert.NotEmpty(t, credentials)
	}).Return(preserveError)

	return credManagerMock
}

func synchronizationServiceMock(applyError error) *kymaMocks.Service {
	syncServiceMock := &kymaMocks.Service{}
	syncServiceMock.On("Apply", kymaModelApps).Return(nil, applyError) // TODO - statuses

	return syncServiceMock
}

func configProviderMock(connectionConfErr, runtimeConfErr error) *configMocks.Provider {
	providerMock := &configMocks.Provider{}
	providerMock.On("GetConnectionConfig").Return(connectionConfig, connectionConfErr)
	providerMock.On("GetRuntimeConfig").Return(runtimeConfig, runtimeConfErr)

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

var (
	compassConnectionNamespacedName = types.NamespacedName{
		Name: compassConnectionName,
	}
)

//
//func TestReconcile(t *testing.T) {
//
//	t.Run("should reconcile request", func(t *testing.T) {
//		// given
//		compassConnection := &v1alpha1.CompassConnection{
//			ObjectMeta: v1.ObjectMeta{
//				Name: compassConnectionName,
//			},
//		}
//
//		fakeClientset := fake.NewSimpleClientset(compassConnection).CompassV1alpha1().CompassConnections()
//		objClient := NewObjectClientWrapper(fakeClientset)
//		supervisor := &mocks.Supervisor{}
//		supervisor.On("SynchronizeWithCompass", compassConnection).Return(compassConnection, nil)
//
//		reconciler := newReconciler(objClient, supervisor, minimalConfigSyncTime)
//
//		request := reconcile.Request{
//			NamespacedName: compassConnectionNamespacedName,
//		}
//
//		// when
//		result, err := reconciler.Reconcile(request)
//
//		// then
//		require.NoError(t, err)
//		require.Empty(t, result)
//
//	})
//
//	t.Run("should reconcile delete request", func(t *testing.T) {
//		// given
//		compassConnection := &v1alpha1.CompassConnection{
//			ObjectMeta: v1.ObjectMeta{
//				Name: compassConnectionName,
//			},
//			Status: v1alpha1.CompassConnectionStatus{
//				State: v1alpha1.Connected,
//			},
//		}
//
//		fakeClientset := fake.NewSimpleClientset().CompassV1alpha1().CompassConnections()
//		objClient := NewObjectClientWrapper(fakeClientset)
//		supervisor := &mocks.Supervisor{}
//		supervisor.On("InitializeCompassConnection").
//			Run(func(args mock.Arguments) {
//				_, err := fakeClientset.Create(compassConnection)
//				require.NoError(t, err)
//			}).
//			Return(compassConnection, nil)
//
//		reconciler := newReconciler(objClient, supervisor, minimalConfigSyncTime)
//
//		request := reconcile.Request{
//			NamespacedName: compassConnectionNamespacedName,
//		}
//
//		// when
//		result, err := reconciler.Reconcile(request)
//
//		// then
//		require.NoError(t, err)
//		require.Empty(t, result)
//
//	})
//}
//
//func Test_shouldResyncConfig(t *testing.T) {
//
//	minimalResyncTime := 300 * time.Second
//
//	for _, testCase := range []struct {
//		description  string
//		syncStatus   *v1alpha1.SynchronizationStatus
//		shouldResync bool
//	}{
//		{
//			description:  "resync if sync status not present",
//			syncStatus:   nil,
//			shouldResync: true,
//		},
//		{
//			description: "resync if sync minimal time passed",
//			syncStatus: &v1alpha1.SynchronizationStatus{
//				LastAttempt: metav1.Unix(time.Now().Unix()-600, 0),
//			},
//			shouldResync: true,
//		},
//		{
//			description: "not resync if sync minimal time did not pass",
//			syncStatus: &v1alpha1.SynchronizationStatus{
//				LastAttempt: metav1.Now(),
//			},
//			shouldResync: false,
//		},
//	} {
//		t.Run("should "+testCase.description, func(t *testing.T) {
//			// given
//			connection := &v1alpha1.CompassConnection{
//				ObjectMeta: v1.ObjectMeta{Name: "connection"},
//				Status: v1alpha1.CompassConnectionStatus{
//					SynchronizationStatus: testCase.syncStatus,
//				},
//			}
//
//			// when
//			resync := shouldResyncConfig(connection, minimalResyncTime)
//
//			// then
//			assert.Equal(t, testCase.shouldResync, resync)
//		})
//	}
//
//}
//
//func NewObjectClientWrapper(client clientset.CompassConnectionInterface) Client {
//	return &objectClientWrapper{
//		fakeClient: client,
//	}
//}
//
//type objectClientWrapper struct {
//	fakeClient clientset.CompassConnectionInterface
//}
//
//func (f *objectClientWrapper) Get(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
//	compassConnection, ok := obj.(*v1alpha1.CompassConnection)
//	if !ok {
//		return errors.New("object is not Compass Connection")
//	}
//
//	cc, err := f.fakeClient.Get(key.Name, v1.GetOptions{})
//	if err != nil {
//		return err
//	}
//
//	cc.DeepCopyInto(compassConnection)
//	return nil
//}
//
//func (f *objectClientWrapper) Update(ctx context.Context, obj runtime.Object, opts ...client.UpdateOption) error {
//	panic("implement me")
//}
//
//func (f *objectClientWrapper) Delete(ctx context.Context, obj runtime.Object, opts ...client.DeleteOption) error {
//	panic("implement me")
//}
