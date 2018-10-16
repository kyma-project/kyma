package testkit

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"testing"

	"github.com/stretchr/testify/require"
)

type ConnectorClient interface {
	CreateToken(t *testing.T) TokenResponse
	GetInfo(t *testing.T, url string) (*InfoResponse, *Error)
	CreateCertChain(t *testing.T, csr, url string) (*CrtResponse, *Error)
}

type connectorClient struct {
	remoteEnv      string
	internalAPIUrl string
	externalAPIUrl string
	httpClient     *http.Client
}

func NewConnectorClient(remoteEnv, internalAPIUrl, externalAPIUrl string, skipVerify bool) ConnectorClient {
	client := NewHttpClient(skipVerify)

	return connectorClient{
		remoteEnv:      remoteEnv,
		internalAPIUrl: internalAPIUrl,
		externalAPIUrl: externalAPIUrl,
		httpClient:     client,
	}
}

func NewHttpClient(skipVerify bool) *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipVerify},
	}
	client := &http.Client{Transport: tr}
	return client
}

func (cc connectorClient) CreateToken(t *testing.T) TokenResponse {
	url := cc.internalAPIUrl + "/v1/remoteenvironments/" + cc.remoteEnv + "/tokens"

	request, err := http.NewRequest(http.MethodPost, url, nil)
	require.NoError(t, err)

	response, err := cc.httpClient.Do(request)
	require.NoError(t, err)
	if response.StatusCode != http.StatusCreated {
		logResponse(t, response)
	}

	require.Equal(t, http.StatusCreated, response.StatusCode)

	tokenResponse := TokenResponse{}

	err = json.NewDecoder(response.Body).Decode(&tokenResponse)
	require.NoError(t, err)

	return tokenResponse
}

func (cc connectorClient) GetInfo(t *testing.T, url string) (*InfoResponse, *Error) {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)

	response, err := cc.httpClient.Do(request)
	require.NoError(t, err)
	if response.StatusCode != http.StatusOK {
		logResponse(t, response)

		errorResponse := ErrorResponse{}
		err = json.NewDecoder(response.Body).Decode(&errorResponse)
		require.NoError(t, err)
		return nil, &Error{response.StatusCode, errorResponse}
	}

	require.Equal(t, http.StatusOK, response.StatusCode)

	infoResponse := &InfoResponse{}

	err = json.NewDecoder(response.Body).Decode(&infoResponse)
	require.NoError(t, err)

	return infoResponse, nil
}

func (cc connectorClient) CreateCertChain(t *testing.T, csr, url string) (*CrtResponse, *Error) {
	body, err := json.Marshal(CsrRequest{Csr: csr})
	require.NoError(t, err)

	request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	require.NoError(t, err)

	request.Header.Add("Content-Type", "application/json")

	response, err := cc.httpClient.Do(request)
	require.NoError(t, err)
	if response.StatusCode != http.StatusCreated {
		logResponse(t, response)
		errorResponse := ErrorResponse{}
		err = json.NewDecoder(response.Body).Decode(&errorResponse)
		require.NoError(t, err)
		return nil, &Error{response.StatusCode, errorResponse}
	}

	require.Equal(t, http.StatusCreated, response.StatusCode)

	crtResponse := &CrtResponse{}

	err = json.NewDecoder(response.Body).Decode(&crtResponse)
	require.NoError(t, err)

	return crtResponse, nil
}

func logResponse(t *testing.T, resp *http.Response) {
	dump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		t.Logf("failed to dump response, %s", err)
	} else {
		t.Logf("\n--------------------------------\n%s\n--------------------------------", dump)
	}
}
