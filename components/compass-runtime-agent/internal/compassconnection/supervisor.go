package compassconnection

import (
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/compass"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/pkg/apis/compass/v1alpha1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO - as a parameter?
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
	InitializeCompassConnectionCR() error
}

func NewSupervisor(connector compass.Connector, crManager CRManager) Supervisor {
	return &crSupervisor{
		compassConnector: connector,
		crManager:        crManager,
		log:              logrus.WithField("Supervisor", "CompassConnection"),
	}
}

type crSupervisor struct {
	compassConnector compass.Connector
	crManager        CRManager
	log              *logrus.Entry
}

func (s *crSupervisor) InitializeCompassConnectionCR() error {
	compassConnectionCR, err := s.crManager.Get(DefaultCompassConnectionName, v1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return s.newCompassConnection()
		}

		return errors.Wrap(err, "Error while getting existing Compass Connection")
	}

	// TODO: Take action based on connection status
	s.log.Infof("Compass Connection exists with state %s", compassConnectionCR.Status.ConnectionState)

	return nil
}

func (s *crSupervisor) newCompassConnection() error {
	connectionCR := &v1alpha1.CompassConnection{
		ObjectMeta: v1.ObjectMeta{
			Name: DefaultCompassConnectionName,
		},
		Spec: v1alpha1.CompassConnectionSpec{},
	}

	err := s.compassConnector.EstablishConnection()
	if err != nil {
		s.log.Errorf("Error while establishing connection with Compass: %s", err.Error())
		s.log.Infof("Creating Compass Connection with NotConnected state")
		connectionCR.Status = v1alpha1.CompassConnectionStatus{
			ConnectionState: v1alpha1.NotConnected,
		}
	} else {
		// TODO: Fill CR Spec with values received from Connection
		connectionCR.Status = v1alpha1.CompassConnectionStatus{
			ConnectionState: v1alpha1.Connected,
		}
	}

	_, err = s.crManager.Create(connectionCR)
	if err != nil {
		return errors.Wrap(err, "Error while creating Compass Connection CR")
	}

	return nil
}
