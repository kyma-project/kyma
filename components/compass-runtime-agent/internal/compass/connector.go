package compass

import (
	"errors"

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
		directorURL: directorURL,
	}
}

type compassConnector struct {
	directorURL string
}

func (cc *compassConnector) EstablishConnection() (EstablishedConnection, error) {
	// TODO: here we should initialize connection with compass

	if cc.directorURL == "" {
		return EstablishedConnection{}, errors.New("Failed to establish connection. Director URL is eempty")
	}

	return EstablishedConnection{
		// TODO: temporary solution until implementation of Connector
		DirectorURL: cc.directorURL,
	}, nil
}
