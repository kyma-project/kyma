package tokenrequest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

const (
	applicationHeader = "Application"
	tenantHeader      = "Tenant"
	groupHeader       = "Group"
	emptyTenant       = ""
	emptyGroup        = ""
)

// TokenDto represents data structure returned from connector-service
type TokenDto struct {
	URL   string `json:"url"`
	Token string `json:"token"`
}

// ConnectorServiceClient interface describes client contract to communicate with connector-service
type ConnectorServiceClient interface {
	FetchToken(appName, tenant, group string) (*TokenDto, error)
}

type connectorServiceClient struct {
	http.Client
	connectorServiceURL string
}

// FetchToken method connects to connector-service and fetches new token for remote-environment
func (c *connectorServiceClient) FetchToken(appName, tenant, group string) (*TokenDto, error) {
	if strings.TrimSpace(appName) == "" {
		return nil, errors.New("appName cannot be empty")
	}

	url := fmt.Sprintf("%s/v1/applications/tokens", c.connectorServiceURL)

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "creating token request")
	}

	req.Header.Set(applicationHeader, appName)

	if tenant != emptyTenant && group != emptyGroup {
		req.Header.Set(tenantHeader, tenant)
		req.Header.Set(groupHeader, group)
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := c.Do(req)
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
