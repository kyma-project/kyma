package services

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/tests/application-connector-tests/test/testkit/connector"
)

const (
	getAllAPIsEndpoint = "/v1/metadata/services"
	sendEventEndpoint  = "/v1/events"
)

type ApplicationConnectorClient struct {
	applicationConnection connector.ApplicationConnection

	httpClient *http.Client
}

func NewApplicationConnectorClient(applicationConnection connector.ApplicationConnection) *ApplicationConnectorClient {
	return &ApplicationConnectorClient{
		applicationConnection: applicationConnection,
		httpClient:            applicationConnection.NewMTLSClient(),
	}
}

func (arc *ApplicationConnectorClient) GetAllAPIs(t *testing.T) (services []Service, errorResponse *ErrorResponse) {
	req, err := http.NewRequest(http.MethodGet, arc.applicationConnection.RegistryURL()+getAllAPIsEndpoint, nil)
	require.NoError(t, err)

	response, err := arc.httpClient.Do(req)
	require.NoError(t, err)

	if response.StatusCode != http.StatusOK {
		err = json.NewDecoder(response.Body).Decode(errorResponse)
		require.NoError(t, err)
		return nil, errorResponse
	}

	err = json.NewDecoder(response.Body).Decode(&services)
	require.NoError(t, err)

	return services, nil
}

func (arc *ApplicationConnectorClient) SendEvent(t *testing.T, eventId string) (publishResponse PublishResponse, errorResponse *ErrorResponse) {
	publishRequest := PublishRequest{
		EventType:        "order.created",
		EventTypeVersion: "v1",
		EventID:          eventId,
		EventTime:        "2012-11-01T22:08:41+00:00",
		Data:             "payload",
	}
	publishRequestEncoded, err := json.Marshal(publishRequest)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, arc.applicationConnection.EventsURL()+sendEventEndpoint, bytes.NewBuffer(publishRequestEncoded))
	require.NoError(t, err)

	response, err := arc.httpClient.Do(req)
	require.NoError(t, err)

	if response.StatusCode != http.StatusOK {
		err = json.NewDecoder(response.Body).Decode(errorResponse)
		return PublishResponse{}, errorResponse
	}

	err = json.NewDecoder(response.Body).Decode(&publishResponse)
	require.NoError(t, err)

	return publishResponse, nil
}
