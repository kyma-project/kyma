package compassconnection

import (
	"fmt"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/compass/director"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/compass/connector"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/certificates"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/pkg/apis/compass/v1alpha1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"

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
	InitializeCompassConnection(token string) (*v1alpha1.CompassConnection, error)
	SynchronizeWithCompass(connection *v1alpha1.CompassConnection) (*v1alpha1.CompassConnection, error)
}

func NewSupervisor(
	connector connector.Connector,
	crManager CRManager,
	credManager certificates.Manager,
	compassClient director.ConfigClient,
	syncService kyma.Service,
) Supervisor {
	return &crSupervisor{
		compassConnector:   connector,
		crManager:          crManager,
		credentialsManager: credManager,
		compassClient:      compassClient,
		syncService:        syncService,
		log:                logrus.WithField("Supervisor", "CompassConnection"),
	}
}

type crSupervisor struct {
	compassConnector   connector.Connector
	crManager          CRManager
	credentialsManager certificates.Manager
	compassClient      director.ConfigClient
	syncService        kyma.Service
	log                *logrus.Entry
}

func (s *crSupervisor) InitializeCompassConnection(token string) (*v1alpha1.CompassConnection, error) {
	compassConnectionCR, err := s.crManager.Get(DefaultCompassConnectionName, v1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return s.newCompassConnection(token)
		}

		return nil, errors.Wrap(err, "Connection failed while getting existing Compass Connection")
	}

	s.log.Infof("Compass Connection exists with state %s", compassConnectionCR.Status.State)

	if !compassConnectionCR.ShouldReconnect() {
		s.log.Infof("Connection already initialized, skipping ")
		return compassConnectionCR, nil
	}

	s.establishConnection(compassConnectionCR, token)

	return s.updateCompassConnection(compassConnectionCR)
}

// SynchronizeWithCompass synchronizes with Compass
func (s *crSupervisor) SynchronizeWithCompass(connection *v1alpha1.CompassConnection) (*v1alpha1.CompassConnection, error) {
	log := s.log.WithField("CompassConnection", connection.Name)
	syncAttemptTime := metav1.Now()

	log.Infof("Getting client credentials...")
	credentials, err := s.credentialsManager.GetClientCredentials()
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to get credentials: %s", err.Error())
		log.Error(errorMsg)
		setSyncFailedStatus(connection, syncAttemptTime, errorMsg)
		return s.updateCompassConnection(connection)
	}

	// TODO: Do we want to call Connector ManagementInfo is it still part of the flow?

	log.Infof("Fetching configuration from Director, from %s url...", connection.Spec.ManagementInfo.DirectorURL)
	applicationsConfig, err := s.compassClient.FetchConfiguration(connection.Spec.ManagementInfo.DirectorURL, credentials)
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to fetch configuration: %s", err.Error())
		log.Error(errorMsg)
		setSyncFailedStatus(connection, syncAttemptTime, errorMsg)
		return s.updateCompassConnection(connection)
	}

	log.Infof("Applying configuration to the cluster...")
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
	log.Infof("Config application results: ")
	for _, res := range results {
		log.Info(res)
	}

	// TODO: report Results to Director

	// TODO: decide the approach of setting this status. Should it be success even if one App failed?
	connection.Status.State = v1alpha1.Synchronized
	connection.Status.SynchronizationStatus = &v1alpha1.SynchronizationStatus{
		LastAttempt:               syncAttemptTime,
		LastSuccessfulFetch:       syncAttemptTime,
		LastSuccessfulApplication: syncAttemptTime,
	}

	connection.Status.State = v1alpha1.Synchronized

	return s.updateCompassConnection(connection)
}

func (s *crSupervisor) newCompassConnection(token string) (*v1alpha1.CompassConnection, error) {
	connectionCR := &v1alpha1.CompassConnection{
		ObjectMeta: v1.ObjectMeta{
			Name: DefaultCompassConnectionName,
		},
		Spec: v1alpha1.CompassConnectionSpec{},
	}

	s.establishConnection(connectionCR, token)

	return s.crManager.Create(connectionCR)
}

func (s *crSupervisor) establishConnection(connectionCR *v1alpha1.CompassConnection, token string) {
	connection, err := s.compassConnector.EstablishConnection(token)
	if err != nil {
		s.log.Errorf("SynchronizationFailed while establishing connection with Compass: %s", err.Error())
		s.log.Infof("Setting Compass Connection to ConnectionFailed state")
		connectionCR.Status.State = v1alpha1.ConnectionFailed
		connectionCR.Status.ConnectionStatus = &v1alpha1.ConnectionStatus{
			LastSync: metav1.Now(),
			Error:    fmt.Sprintf("Failed to retrieve certificate: %s", err.Error()),
		}
	} else {
		s.saveCredentials(connectionCR, connection)
	}
}

func (s *crSupervisor) saveCredentials(connectionCR *v1alpha1.CompassConnection, establishedConnection connector.EstablishedConnection) {
	err := s.credentialsManager.PreserveCredentials(establishedConnection.Credentials)
	if err != nil {
		s.log.Errorf("SynchronizationFailed while preserving credentials: %s", err.Error())
		s.log.Infof("Setting Compass Connection to ConnectionFailed state")
		connectionCR.Status.State = v1alpha1.ConnectionFailed
		connectionCR.Status.ConnectionStatus = &v1alpha1.ConnectionStatus{
			Error: fmt.Sprintf("Failed to preserve certificate: %s", err.Error()),
		}

		return
	}

	s.log.Infof("Connection established. Director URL: %s", establishedConnection.DirectorURL)

	connectionTime := metav1.Now()

	// TODO: set certificate status when we will have full flow
	connectionCR.Status.State = v1alpha1.Connected
	connectionCR.Status.ConnectionStatus = &v1alpha1.ConnectionStatus{
		Established: connectionTime,
		LastSync:    connectionTime,
		LastSuccess: connectionTime,
	}

	connectionCR.Spec.ManagementInfo = v1alpha1.ManagementInfo{
		DirectorURL: establishedConnection.DirectorURL,
	}
}

func (s *crSupervisor) updateCompassConnection(connectionCR *v1alpha1.CompassConnection) (*v1alpha1.CompassConnection, error) {
	// TODO: with retries

	return s.crManager.Update(connectionCR)
}

func setSyncFailedStatus(connection *v1alpha1.CompassConnection, attemptTime metav1.Time, errorMsg string) {
	connection.Status.State = v1alpha1.SynchronizationFailed
	connection.Status.SynchronizationStatus = &v1alpha1.SynchronizationStatus{
		LastAttempt: attemptTime,
		Error:       errorMsg,
	}
}
