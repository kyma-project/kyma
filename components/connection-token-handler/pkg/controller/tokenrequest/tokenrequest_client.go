package tokenrequest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

// TokenDto represents data structure returned from connector-service
type TokenDto struct {
	URL   string `json:"url"`
	Token string `json:"token"`
}

// ConnectorServiceClient interface describes client contract to communicate with connector-service
type ConnectorServiceClient interface {
	FetchToken(reName string) (*TokenDto, error)
}

type connectorServiceClient struct {
	http.Client
	connectorServiceURL string
}

// FetchToken method connects to connector-service and fetches new token for remote-environment
func (c *connectorServiceClient) FetchToken(reName string) (*TokenDto, error) {
	if strings.TrimSpace(reName) == "" {
		return nil, errors.New("reName cannot be empty")
	}

	url := fmt.Sprintf("%s/v1/remoteenvironments/%s/tokens", c.connectorServiceURL, reName)

	res, err := c.Post(url, "application/json", nil)
	if err != nil {
		return nil, errors.Wrap(err, "while issuing POST request")
	}

	defer res.Body.Close()
	token := &TokenDto{}
	if err := json.NewDecoder(res.Body).Decode(token); err != nil {
		return nil, errors.Wrap(err, "while decoding json")
	}

	return token, nil
}

// NewConnectorServiceClient constucts new instance of connector service client
func NewConnectorServiceClient(connectorServiceURL string) ConnectorServiceClient {
	return &connectorServiceClient{
		connectorServiceURL: connectorServiceURL,
	}
}
