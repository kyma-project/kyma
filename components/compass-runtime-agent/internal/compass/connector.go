package compass

import (
	"encoding/json"
	"io/ioutil"

	"github.com/pkg/errors"
)

type TokenURL struct {
	URL   string `json:"url"`
	Token string `json:"token"`
}

//go:generate mockery -name=Connector
type Connector interface {
	// TODO: we should return certificates and management URLs
	EstablishConnection() error
}

func NewCompassConnector(tokenURLFile string) Connector {
	return &compassConnector{
		tokenURLFile: tokenURLFile,
	}
}

type compassConnector struct {
	tokenURLFile string
}

func (cc *compassConnector) EstablishConnection() error {
	data, err := ioutil.ReadFile(cc.tokenURLFile)
	if err != nil {
		return errors.Wrap(err, "Failed to read token URL file")
	}

	var tokenURL TokenURL
	err = json.Unmarshal(data, &tokenURL)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal token URL data")
	}

	// TODO: here we should initialize connection with compass

	return nil
}
