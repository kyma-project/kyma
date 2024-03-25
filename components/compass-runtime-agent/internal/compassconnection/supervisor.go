package compassconnection

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"time"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/compass/cache"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/compass/director"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/compass"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/config"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/certificates"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/pkg/apis/compass/v1alpha1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	DefaultCompassConnectionName = "compass-connection"
)

//go:generate mockery --name=CRManager
type CRManager interface {
	Create(ctx context.Context, cc *v1alpha1.CompassConnection, options metav1.CreateOptions) (*v1alpha1.CompassConnection, error)
	Update(ctx context.Context, cc *v1alpha1.CompassConnection, options metav1.UpdateOptions) (*v1alpha1.CompassConnection, error)
	Delete(ctx context.Context, name string, options metav1.DeleteOptions) error
	Get(ctx context.Context, name string, options metav1.GetOptions) (*v1alpha1.CompassConnection, error)
}

//go:generate mockery --name=Supervisor
type Supervisor interface {
	InitializeCompassConnection(ctx context.Context) (*v1alpha1.CompassConnection, error)
	SynchronizeWithCompass(ctx context.Context, connection *v1alpha1.CompassConnection) (*v1alpha1.CompassConnection, error)
	MaintainCompassConnection(ctx context.Context, connection *v1alpha1.CompassConnection) error
}

func NewSupervisor(
	connector Connector,
	crManager CRManager,
	credManager certificates.Manager,
	clientsProvider compass.ClientsProvider,
	syncService kyma.Service,
	configProvider config.Provider,
	certValidityRenewalThreshold float64,
	minimalCompassSyncTime time.Duration,
	runtimeURLsConfig director.RuntimeURLsConfig,
	connectionDataCache cache.ConnectionDataCache,
) Supervisor {
	return &crSupervisor{
		compassConnector:             connector,
		crManager:                    crManager,
		credentialsManager:           credManager,
		clientsProvider:              clientsProvider,
		syncService:                  syncService,
		configProvider:               configProvider,
		certValidityRenewalThreshold: certValidityRenewalThreshold,
		minimalCompassSyncTime:       minimalCompassSyncTime,
		runtimeURLsConfig:            runtimeURLsConfig,
		connectionDataCache:          connectionDataCache,
		log:                          logrus.WithField("Supervisor", "CompassConnection"),
	}
}

type crSupervisor struct {
	compassConnector             Connector
	crManager                    CRManager
	credentialsManager           certificates.Manager
	clientsProvider              compass.ClientsProvider
	syncService                  kyma.Service
	configProvider               config.Provider
	certValidityRenewalThreshold float64
	minimalCompassSyncTime       time.Duration
	runtimeURLsConfig            director.RuntimeURLsConfig
	log                          *logrus.Entry
	connectionDataCache          cache.ConnectionDataCache
}

func (s *crSupervisor) InitializeCompassConnection(ctx context.Context) (*v1alpha1.CompassConnection, error) {
	compassConnectionCR, err := s.crManager.Get(context.Background(), DefaultCompassConnectionName, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return s.newCompassConnection(ctx)
		}

		return nil, errors.Wrap(err, "Connection failed while getting existing Compass Connection")
	}

	s.log.Infof("Compass Connection exists with state %s", compassConnectionCR.Status.State)

	if !compassConnectionCR.Failed() {
		s.log.Infof("Connection already initialized, skipping ")

		credentials, err := s.credentialsManager.GetClientCredentials()
		if err != nil {
			return nil, fmt.Errorf("failed to read credentials while initializing Compass Connection CR: %s", err.Error())
		}

		s.connectionDataCache.UpdateConnectionData(
			credentials.AsTLSCertificate(),
			compassConnectionCR.Spec.ManagementInfo.DirectorURL,
			compassConnectionCR.Spec.ManagementInfo.ConnectorURL,
		)

		return compassConnectionCR, nil
	}

	s.establishConnection(ctx, compassConnectionCR)

	return s.updateCompassConnection(compassConnectionCR)
}

func (s *crSupervisor) MaintainCompassConnection(ctx context.Context, connection *v1alpha1.CompassConnection) error {
	s.log = s.log.WithField("CompassConnection", connection.Name)

	s.log.Infof("Trying to maintain connection to Connector with %s url...", connection.Spec.ManagementInfo.ConnectorURL)
	err := s.maintainCompassConnection(ctx, connection)
	if err != nil {
		errorMsg := fmt.Sprintf("Error while trying to maintain connection: %s", err.Error())
		s.setConnectionMaintenanceFailedStatus(connection, metav1.Now(), errorMsg) // save in ConnectionStatus.LastSync
		_, updateErr := s.updateCompassConnection(connection)

		if updateErr != nil {
			return updateErr
		}
	}
	return err
}

func (s *crSupervisor) SynchronizeWithCompass(ctx context.Context, connection *v1alpha1.CompassConnection) (*v1alpha1.CompassConnection, error) {
	s.log = s.log.WithField("CompassConnection", connection.Name)

	s.log.Infof("Reading configuration required to fetch Runtime configuration...")
	runtimeConfig, err := s.configProvider.GetRuntimeConfig()
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to read Runtime config: %s", err.Error())
		s.setSyncFailedStatus(connection, metav1.Now(), errorMsg) // save in SynchronizationStatus.LastAttempt
		return s.updateCompassConnection(connection)
	}

	s.log.Infof("Fetching configuration from Director, from %s url...", connection.Spec.ManagementInfo.DirectorURL)
	directorClient, err := s.clientsProvider.GetDirectorClient(runtimeConfig)
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to prepare configuration client: %s", err.Error())
		s.setSyncFailedStatus(connection, metav1.Now(), errorMsg) // save in SynchronizationStatus.LastAttempt
		return s.updateCompassConnection(connection)
	}

	s.log.Info("Checking whether Runtime has CONNECTED status in Compass")
	if !s.runtimeHasConnectedStatusInCompass(connection) {
		s.log.Info("Setting CONNECTED status for Runtime")
		err := directorClient.SetRuntimeStatusCondition(ctx, graphql.RuntimeStatusConditionConnected)
		if err != nil {
			return connection, err
		}

		s.setRuntimeStatusInCompass(connection)

		// <AG> Just for testing
		runtime, err := directorClient.GetRuntime(ctx)
		if err != nil {
			s.log.Infof("TEST: Current Runtime status: %s", runtime.Status.Condition.String())
		}
		// <AG>
	}

	applicationsConfig, runtimeLabels, err := directorClient.FetchConfiguration(ctx)
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to fetch configuration: %s", err.Error())
		s.setSyncFailedStatus(connection, metav1.Now(), errorMsg) // save in SynchronizationStatus.LastAttempt
		return s.updateCompassConnection(connection)
	}

	s.log.Infof("Applying configuration to the cluster...")
	results, err := s.syncService.Apply(applicationsConfig)
	if err != nil {
		syncAttemptTime := metav1.Now()
		connection.Status.State = v1alpha1.ResourceApplicationFailed
		connection.Status.SynchronizationStatus = &v1alpha1.SynchronizationStatus{
			LastAttempt:         syncAttemptTime,
			LastSuccessfulFetch: syncAttemptTime,
			Error:               fmt.Sprintf("Failed to apply configuration: %s", err.Error()),
		}
		return s.updateCompassConnection(connection)
	}

	// TODO: save result to CR and possibly log in better manner
	s.log.Infof("Config application results: ")
	for _, res := range results {
		s.log.Info(res)
	}

	s.log.Infof("Labeling Runtime with URLs...")
	_, err = directorClient.SetURLsLabels(ctx, s.runtimeURLsConfig, runtimeLabels)
	if err != nil {
		syncAttemptTime := metav1.Now()
		connection.Status.State = v1alpha1.MetadataUpdateFailed
		connection.Status.SynchronizationStatus = &v1alpha1.SynchronizationStatus{
			LastAttempt:         syncAttemptTime,
			LastSuccessfulFetch: syncAttemptTime,
			Error:               fmt.Sprintf("Failed to reconcile Runtime labels with proper URLs: %s", err.Error()),
		}
		return s.updateCompassConnection(connection)
	}

	// TODO: decide the approach of setting this status. Should it be success even if one App failed?
	s.setConnectionSynchronizedStatus(connection, metav1.Now())
	connection.Spec.ResyncNow = false

	return s.updateCompassConnection(connection)
}

func (s *crSupervisor) runtimeHasConnectedStatusInCompass(compassConnection *v1alpha1.CompassConnection) bool {
	annotations := compassConnection.Annotations
	_, found := annotations["operator.kyma-project.io/compass-status-connected"]

	return found
}

func (s *crSupervisor) setRuntimeStatusInCompass(compassConnection *v1alpha1.CompassConnection) {
	annotations := compassConnection.Annotations

	if annotations == nil {
		annotations = map[string]string{}
	}

	annotations["operator.kyma-project.io/compass-status-connected"] = string(graphql.RuntimeStatusConditionConnected)
	compassConnection.Annotations = annotations
}

func (s *crSupervisor) maintainCompassConnection(ctx context.Context, compassConnection *v1alpha1.CompassConnection) error {
	shouldRenew := compassConnection.ShouldRenewCertificate(s.certValidityRenewalThreshold, s.minimalCompassSyncTime)
	credentialsExist, err := s.credentialsManager.CredentialsExist()
	if err != nil {
		return errors.Wrap(err, "Failed to check whether credentials exist")
	}

	s.log.Infof("Trying to maintain certificates connection... Renewal: %v, CreadentialsExist: %v", shouldRenew, credentialsExist)
	newCreds, managementInfo, err := s.compassConnector.MaintainConnection(ctx, shouldRenew, credentialsExist)
	if err != nil {
		return errors.Wrap(err, "Failed to connect to Compass Connector")
	}

	connectionTime := metav1.Now()

	if newCreds != nil {
		s.log.Infof("Trying to save renewed certificates...")
		err = s.credentialsManager.PreserveCredentials(*newCreds)
		if err != nil {
			return errors.Wrap(err, "Failed to preserve certificate")
		}

		s.log.Infof("Successfully saved renewed certificates")
		compassConnection.SetCertificateStatus(connectionTime, newCreds.ClientCertificate)
		compassConnection.Spec.RefreshCredentialsNow = false
		compassConnection.Status.ConnectionStatus.Renewed = connectionTime

		s.connectionDataCache.UpdateConnectionData((*newCreds).AsTLSCertificate(), managementInfo.DirectorURL, managementInfo.ConnectorURL)
		s.log.Infof("Refreshed connection data cache")
	}

	if s.urlsUpdated(compassConnection, managementInfo) {
		s.log.Infof("Compass URLs modified. Updating cache. Connector: %s => %s, Director: %s => %s",
			compassConnection.Spec.ManagementInfo.ConnectorURL, managementInfo.ConnectorURL,
			compassConnection.Spec.ManagementInfo.DirectorURL, managementInfo.DirectorURL)
		s.connectionDataCache.UpdateURLs(managementInfo.DirectorURL, managementInfo.ConnectorURL)
	}

	s.log.Infof("Connection maintained. Director URL: %s , ConnectorURL: %s", managementInfo.DirectorURL, managementInfo.ConnectorURL)

	if compassConnection.Status.ConnectionStatus == nil {
		compassConnection.Status.ConnectionStatus = &v1alpha1.ConnectionStatus{}
	}

	compassConnection.Status.ConnectionStatus.LastSync = connectionTime
	compassConnection.Status.ConnectionStatus.LastSuccess = connectionTime
	// for alternative flow compassConnection.Status.State = v1alpha1.Connected

	return nil
}

func (s *crSupervisor) urlsUpdated(compassConnectionCR *v1alpha1.CompassConnection, managementInfo v1alpha1.ManagementInfo) bool {
	return compassConnectionCR.Spec.ManagementInfo.ConnectorURL != managementInfo.ConnectorURL ||
		compassConnectionCR.Spec.ManagementInfo.DirectorURL != managementInfo.DirectorURL
}

func (s *crSupervisor) newCompassConnection(ctx context.Context) (*v1alpha1.CompassConnection, error) {
	connectionCR := &v1alpha1.CompassConnection{
		ObjectMeta: metav1.ObjectMeta{
			Name: DefaultCompassConnectionName,
		},
		Spec: v1alpha1.CompassConnectionSpec{},
	}

	s.establishConnection(ctx, connectionCR)

	return s.crManager.Create(context.Background(), connectionCR, metav1.CreateOptions{})
}

func (s *crSupervisor) establishConnection(ctx context.Context, connectionCR *v1alpha1.CompassConnection) {
	connCfg, err := s.configProvider.GetConnectionConfig()
	if err != nil {
		s.setConnectionFailedStatus(connectionCR, err, fmt.Sprintf("Failed to retrieve certificate: %s", err.Error()))
		return
	}

	connection, err := s.compassConnector.EstablishConnection(ctx, connCfg.ConnectorURL, connCfg.Token)
	if err != nil {
		s.setConnectionFailedStatus(connectionCR, err, fmt.Sprintf("Failed to retrieve certificate: %s", err.Error()))
		return
	}

	connectionTime := metav1.Now()

	err = s.credentialsManager.PreserveCredentials(connection.Credentials)
	if err != nil {
		s.setConnectionFailedStatus(connectionCR, err, fmt.Sprintf("Failed to preserve certificate: %s", err.Error()))
		return
	}

	s.log.Infof("Connection established. Director URL: %s , ConnectorURL: %s", connection.ManagementInfo.DirectorURL, connection.ManagementInfo.ConnectorURL)

	connectionCR.Status.State = v1alpha1.Connected
	connectionCR.Status.ConnectionStatus = &v1alpha1.ConnectionStatus{
		Established: connectionTime,
		LastSync:    connectionTime,
		LastSuccess: connectionTime,
	}
	connectionCR.SetCertificateStatus(connectionTime, connection.Credentials.ClientCertificate)

	connectionCR.Spec.ManagementInfo = connection.ManagementInfo

	s.connectionDataCache.UpdateConnectionData(
		connection.Credentials.AsTLSCertificate(),
		connection.ManagementInfo.DirectorURL,
		connection.ManagementInfo.ConnectorURL,
	)
}

func (s *crSupervisor) setConnectionFailedStatus(connectionCR *v1alpha1.CompassConnection, err error, connStatusError string) {
	s.log.Errorf("Error while establishing connection with Compass: %s", err.Error())
	s.log.Infof("Setting Compass Connection to ConnectionFailed state")
	connectionCR.Status.State = v1alpha1.ConnectionFailed
	if connectionCR.Status.ConnectionStatus == nil {
		connectionCR.Status.ConnectionStatus = &v1alpha1.ConnectionStatus{}
	}
	connectionCR.Status.ConnectionStatus.LastSync = metav1.Now()
	connectionCR.Status.ConnectionStatus.Error = connStatusError
}

func (s *crSupervisor) setConnectionSynchronizedStatus(connectionCR *v1alpha1.CompassConnection, attemptTime metav1.Time) {
	s.log.Infof("Setting Compass Connection to Synchronized state")
	connectionCR.Status.State = v1alpha1.Synchronized
	connectionCR.Status.SynchronizationStatus = &v1alpha1.SynchronizationStatus{
		LastAttempt:               attemptTime,
		LastSuccessfulFetch:       attemptTime,
		LastSuccessfulApplication: attemptTime,
	}
}

func (s *crSupervisor) setConnectionMaintenanceFailedStatus(connectionCR *v1alpha1.CompassConnection, attemptTime metav1.Time, errorMsg string) {
	s.log.Error(errorMsg)
	s.log.Infof("Setting Compass Connection to ConnectionMaintenanceFailed state")
	connectionCR.Status.State = v1alpha1.ConnectionMaintenanceFailed
	if connectionCR.Status.ConnectionStatus == nil {
		connectionCR.Status.ConnectionStatus = &v1alpha1.ConnectionStatus{}
	}
	connectionCR.Status.ConnectionStatus.LastSync = attemptTime
	connectionCR.Status.ConnectionStatus.Error = errorMsg
}

func (s *crSupervisor) updateCompassConnection(connectionCR *v1alpha1.CompassConnection) (*v1alpha1.CompassConnection, error) {
	// TODO: with retries

	return s.crManager.Update(context.Background(), connectionCR, metav1.UpdateOptions{})
}

func (s *crSupervisor) setSyncFailedStatus(connectionCR *v1alpha1.CompassConnection, attemptTime metav1.Time, errorMsg string) {
	s.log.Error(errorMsg)
	s.log.Infof("Setting Compass Connection to SynchronizationFailed state")
	connectionCR.Status.State = v1alpha1.SynchronizationFailed
	if connectionCR.Status.SynchronizationStatus == nil {
		connectionCR.Status.SynchronizationStatus = &v1alpha1.SynchronizationStatus{}
	}
	connectionCR.Status.SynchronizationStatus.LastAttempt = attemptTime
	connectionCR.Status.SynchronizationStatus.Error = errorMsg
}
