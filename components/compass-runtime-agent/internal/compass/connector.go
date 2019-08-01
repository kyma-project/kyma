package compass

import (
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/certificates"
)

type EstablishedConnection struct {
	Credentials certificates.Credentials
	DirectorURL string
}

//go:generate mockery -name=Connector
type Connector interface {
	EstablishConnection() (EstablishedConnection, error)
}

func NewCompassConnector(directorURL string) Connector {
	return &compassConnector{
		directorULR: directorURL,
	}
}

type compassConnector struct {
	directorULR string
}

func (cc *compassConnector) EstablishConnection() (EstablishedConnection, error) {
	// TODO: here we should initialize connection with compass

	return EstablishedConnection{
		// TODO: temporary solution until implementation of Connector
		DirectorURL: cc.directorULR,
	}, nil
}
