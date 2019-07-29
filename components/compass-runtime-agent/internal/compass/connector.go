package compass

import (
	"encoding/json"
	"io/ioutil"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/certificates"

	"github.com/pkg/errors"
)

type EstablishedConnection struct {
	Credentials certificates.Credentials
	DirectorURL string
}

type TokenURL struct {
	URL   string `json:"url"`
	Token string `json:"token"`
}

//go:generate mockery -name=Connector
type Connector interface {
	// TODO: we should return certificates and management URLs
	EstablishConnection() (EstablishedConnection, error)
}

func NewCompassConnector(tokenURLFile string) Connector {
	return &compassConnector{
		tokenURLFile: tokenURLFile,
	}
}

type compassConnector struct {
	tokenURLFile string
}

func (cc *compassConnector) EstablishConnection() (EstablishedConnection, error) {
	data, err := ioutil.ReadFile(cc.tokenURLFile)
	if err != nil {
		return EstablishedConnection{}, errors.Wrap(err, "Failed to read token URL file")
	}

	var tokenURL TokenURL
	err = json.Unmarshal(data, &tokenURL)
	if err != nil {
		return EstablishedConnection{}, errors.Wrap(err, "Failed to unmarshal token URL data")
	}

	// TODO: here we should initialize connection with compass

	// TODO: here we should return EstablishedConnection with certificates and URL
	return EstablishedConnection{
		// TODO: temporary solution until implementation of Connector
		// For now Director URL is passed directly instead of TokenURL
		DirectorURL: tokenURL.URL,
	}, nil
}
