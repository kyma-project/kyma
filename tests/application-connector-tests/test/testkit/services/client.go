package services

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/tests/application-connector-tests/test/testkit/connector"
	"github.com/kyma-project/kyma/tests/application-connector-tests/test/testkit/util"
)

type ApplicationConnectorClient struct {
	applicationCredentials connector.ApplicationCredentials
	managementURLs         connector.ManagementInfoURLs

	httpClient *http.Client
}

func NewApplicationConnectorClient(credentials connector.ApplicationCredentials, urls connector.ManagementInfoURLs, skipSSLVerify bool) *ApplicationConnectorClient {
	return &ApplicationConnectorClient{
		applicationCredentials: credentials,
		managementURLs:         urls,
		httpClient:             credentials.NewMTLSClient(skipSSLVerify),
	}
}

func (arc *ApplicationConnectorClient) GetAllAPIs(t *testing.T) ([]Service, *ErrorResponse) {
	req, err := http.NewRequest(http.MethodGet, arc.managementURLs.MetadataUrl, nil)
	require.NoError(t, err)

	response, err := arc.httpClient.Do(req)
	require.NoError(t, err)
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		errorResponse := handleErrorResponse(t, response)
		return nil, &errorResponse
	}

	var services []Service
	err = json.NewDecoder(response.Body).Decode(&services)
	require.NoError(t, err)

	return services, nil
}

func (arc *ApplicationConnectorClient) SendEvent(t *testing.T, eventId string) (PublishResponse, *ErrorResponse) {
	publishRequest := PublishRequest{
		EventType:        "order.created",
		EventTypeVersion: "v1",
		EventID:          eventId,
		EventTime:        "2012-11-01T22:08:41+00:00",
		Data:             "payload",
	}
	publishRequestEncoded, err := json.Marshal(publishRequest)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, arc.managementURLs.EventsUrl, bytes.NewBuffer(publishRequestEncoded))
	require.NoError(t, err)

	response, err := arc.httpClient.Do(req)
	require.NoError(t, err)
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		errorResponse := handleErrorResponse(t, response)
		return PublishResponse{}, &errorResponse
	}

	var publishResponse PublishResponse
	err = json.NewDecoder(response.Body).Decode(&publishResponse)
	require.NoError(t, err)

	return publishResponse, nil
}

func handleErrorResponse(t *testing.T, response *http.Response) ErrorResponse {
	util.LogResponse(t, response)

	var errorResponse ErrorResponse
	err := json.NewDecoder(response.Body).Decode(&errorResponse)
	if err != nil {
		errData, err := ioutil.ReadAll(response.Body)
		require.NoError(t, err)

		errorResponse = ErrorResponse{
			Code:  response.StatusCode,
			Error: string(errData),
		}
	}
	return errorResponse
}

func RequireNoError(t *testing.T, errorResponse *ErrorResponse) {
	require.Nil(t, errorResponse)
}
