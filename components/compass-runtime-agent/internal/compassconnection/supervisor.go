package compassconnection

import (
	"fmt"
	"time"

	"kyma-project.io/compass-runtime-agent/internal/compass/director"

	"kyma-project.io/compass-runtime-agent/internal/compass"

	"kyma-project.io/compass-runtime-agent/internal/config"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"kyma-project.io/compass-runtime-agent/internal/certificates"
	"kyma-project.io/compass-runtime-agent/internal/kyma"
	"kyma-project.io/compass-runtime-agent/pkg/apis/compass/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	DefaultCompassConnectionName = "compass-connection"
)

//go:generate mockery -name=CRManager
type CRManager interface {
	Create(*v1alpha1.CompassConnection) (*v1alpha1.CompassConnection, error)
	Update(*v1alpha1.CompassConnection) (*v1alpha1.CompassConnection, error)
	Delete(name string, options *v1.DeleteOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.CompassConnection, error)
}

//go:generate mockery -name=Supervisor
type Supervisor interface {
	InitializeCompassConnection() (*v1alpha1.CompassConnection, error)
	SynchronizeWithCompass(connection *v1alpha1.CompassConnection) (*v1alpha1.CompassConnection, error)
}

func NewSupervisor(
	connector Connector,
	crManager CRManager,
	credManager certificates.Manager,
	clientsProvider compass.ClientsProvider,
	syncService kyma.Service,
	configProvider config.Provider,
	directorProxyConfigurator director.ProxyConfigurator,
	certValidityRenewalThreshold float64,
	minimalCompassSyncTime time.Duration,
	runtimeURLsConfig director.RuntimeURLsConfig,
) Supervisor {
	return &crSupervisor{
		compassConnector:             connector,
		crManager:                    crManager,
		credentialsManager:           credManager,
		clientsProvider:              clientsProvider,
		syncService:                  syncService,
		configProvider:               configProvider,
		directorProxyConfigurator:    directorProxyConfigurator,
		certValidityRenewalThreshold: certValidityRenewalThreshold,
		minimalCompassSyncTime:       minimalCompassSyncTime,
		runtimeURLsConfig:            runtimeURLsConfig,
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
	directorProxyConfigurator    director.ProxyConfigurator
}

func (s *crSupervisor) InitializeCompassConnection() (*v1alpha1.CompassConnection, error) {
	compassConnectionCR, err := s.crManager.Get(DefaultCompassConnectionName, v1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return s.newCompassConnection()
		}

		return nil, errors.Wrap(err, "Connection failed while getting existing Compass Connection")
	}

	s.log.Infof("Compass Connection exists with state %s", compassConnectionCR.Status.State)

	if !compassConnectionCR.ShouldAttemptReconnect() {
		s.log.Infof("Connection already initialized, skipping ")
		return compassConnectionCR, nil
	}

	s.establishConnection(compassConnectionCR)

	return s.updateCompassConnection(compassConnectionCR)
}

// SynchronizeWithCompass synchronizes with Compass
func (s *crSupervisor) SynchronizeWithCompass(connection *v1alpha1.CompassConnection) (*v1alpha1.CompassConnection, error) {
	s.log = s.log.WithField("CompassConnection", connection.Name)
	syncAttemptTime := metav1.Now()

	s.log.Infof("Getting client credentials...")
	credentials, err := s.credentialsManager.GetClientCredentials()
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to get client credentials: %s", err.Error())
		s.setConnectionMaintenanceFailedStatus(connection, syncAttemptTime, errorMsg)
		return s.updateCompassConnection(connection)
	}

	s.log.Infof("Trying to maintain connection to Connector with %s url...", connection.Spec.ManagementInfo.ConnectorURL)
	err = s.maintainCompassConnection(credentials, connection)
	if err != nil {
		errorMsg := fmt.Sprintf("Error while trying to maintain connection: %s", err.Error())
		s.setConnectionMaintenanceFailedStatus(connection, syncAttemptTime, errorMsg)
		return s.updateCompassConnection(connection)
	}

	s.log.Infof("Updating Director proxy configuration...")
	err = s.directorProxyConfigurator.SetURLAndCerts(connection.Spec.ManagementInfo.DirectorURL, credentials.AsTLSCertificate())
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to update Director proxy configuration: %s", err.Error())
		s.setSyncFailedStatus(connection, syncAttemptTime, errorMsg)
		return s.updateCompassConnection(connection)
	}

	s.log.Infof("Reading configuration required to fetch Runtime configuration...")
	runtimeConfig, err := s.configProvider.GetRuntimeConfig()
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to read Runtime config: %s", err.Error())
		s.setSyncFailedStatus(connection, syncAttemptTime, errorMsg)
		return s.updateCompassConnection(connection)
	}

	s.log.Infof("Fetching configuration from Director, from %s url...", connection.Spec.ManagementInfo.DirectorURL)
	directorClient, err := s.clientsProvider.GetDirectorClient(credentials, connection.Spec.ManagementInfo.DirectorURL, runtimeConfig)
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to prepare configuration client: %s", err.Error())
		s.setSyncFailedStatus(connection, syncAttemptTime, errorMsg)
		return s.updateCompassConnection(connection)
	}

	applicationsConfig, err := directorClient.FetchConfiguration()
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to fetch configuration: %s", err.Error())
		s.setSyncFailedStatus(connection, syncAttemptTime, errorMsg)
		return s.updateCompassConnection(connection)
	}

	s.log.Infof("Applying configuration to the cluster...")
	results, err := s.syncService.Apply(applicationsConfig)
	if err != nil {
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
	_, err = directorClient.SetURLsLabels(s.runtimeURLsConfig)
	if err != nil {
		connection.Status.State = v1alpha1.MetadataUpdateFailed
		connection.Status.SynchronizationStatus = &v1alpha1.SynchronizationStatus{
			LastAttempt:         syncAttemptTime,
			LastSuccessfulFetch: syncAttemptTime,
			Error:               fmt.Sprintf("Failed to label Runtime with proper URLs: %s", err.Error()),
		}
		return s.updateCompassConnection(connection)
	}

	// TODO: decide the approach of setting this status. Should it be success even if one App failed?
	s.setConnectionSynchronizedStatus(connection, syncAttemptTime)
	connection.Spec.ResyncNow = false

	return s.updateCompassConnection(connection)
}

func (s *crSupervisor) maintainCompassConnection(credentials certificates.ClientCredentials, compassConnection *v1alpha1.CompassConnection) error {
	shouldRenew := compassConnection.ShouldRenewCertificate(s.certValidityRenewalThreshold, s.minimalCompassSyncTime)

	s.log.Infof("Trying to maintain certificates connection... Renewal: %v", shouldRenew)
	newCreds, managementInfo, err := s.compassConnector.MaintainConnection(credentials, compassConnection.Spec.ManagementInfo.ConnectorURL, shouldRenew)
	if err != nil {
		return errors.Wrap(err, "Failed to connect to Compass Connector")
	}

	connectionTime := metav1.Now()

	if shouldRenew && newCreds != nil {
		s.log.Infof("Trying to save renewed certificates...")
		err = s.credentialsManager.PreserveCredentials(*newCreds)
		if err != nil {
			return errors.Wrap(err, "Failed to preserve certificate")
		}

		s.log.Infof("Successfully saved renewed certificates")
		compassConnection.SetCertificateStatus(connectionTime, newCreds.ClientCertificate)
		compassConnection.Spec.RefreshCredentialsNow = false
		compassConnection.Status.ConnectionStatus.Renewed = connectionTime
	}

	s.log.Infof("Connection maintained. Director URL: %s , ConnectorURL: %s", managementInfo.DirectorURL, managementInfo.ConnectorURL)

	if compassConnection.Status.ConnectionStatus == nil {
		compassConnection.Status.ConnectionStatus = &v1alpha1.ConnectionStatus{}
	}

	compassConnection.Status.ConnectionStatus.LastSync = connectionTime
	compassConnection.Status.ConnectionStatus.LastSuccess = connectionTime
	compassConnection.Spec.ManagementInfo = managementInfo

	return nil
}

func (s *crSupervisor) newCompassConnection() (*v1alpha1.CompassConnection, error) {
	connectionCR := &v1alpha1.CompassConnection{
		ObjectMeta: v1.ObjectMeta{
			Name: DefaultCompassConnectionName,
		},
		Spec: v1alpha1.CompassConnectionSpec{},
	}

	s.establishConnection(connectionCR)

	return s.crManager.Create(connectionCR)
}

func (s *crSupervisor) establishConnection(connectionCR *v1alpha1.CompassConnection) {
	connCfg, err := s.configProvider.GetConnectionConfig()
	if err != nil {
		s.setConnectionFailedStatus(connectionCR, err, fmt.Sprintf("Failed to retrieve certificate: %s", err.Error()))
		return
	}

	connection, err := s.compassConnector.EstablishConnection(connCfg.ConnectorURL, connCfg.Token)
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

	return s.crManager.Update(connectionCR)
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
