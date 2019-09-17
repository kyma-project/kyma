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

func NewCompassConnector(connectorURL string, csrProvider certificates.CSRProvider) Connector {
	return &compassConnector{
		connectorURL: connectorURL,
	}
}

type compassConnector struct {
	connectorURL string
	csrProvider  certificates.CSRProvider
}

func (cc *compassConnector) EstablishConnection() (EstablishedConnection, error) {
	// TODO: here we should initialize connection with compass

	if cc.connectorURL == "" {
		return EstablishedConnection{}, errors.New("Failed to establish connection. Connector URL is eempty")
	}

	// TODO - fetch config

	return EstablishedConnection{
		// TODO: temporary solution until implementation of Connector
		DirectorURL: cc.connectorURL,
	}, nil
}
