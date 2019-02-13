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

const (
	ApplicationHeader = "Application"
	GroupHeader       = "Group"
	TenantHeader      = "Tenant"
)

type ConnectorClient interface {
	CreateToken(t *testing.T) TokenResponse
	GetInfo(t *testing.T, url string, headers map[string]string) (*InfoResponse, *Error)
	CreateCertChain(t *testing.T, csr, url string) (*CrtResponse, *Error)
}

type connectorClient struct {
	httpClient   *http.Client
	tokenRequest *http.Request
}

func NewConnectorClient(tokenRequest *http.Request, skipVerify bool) ConnectorClient {
	client := NewHttpClient(skipVerify)

	return connectorClient{
		httpClient:   client,
		tokenRequest: tokenRequest,
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
	response, err := cc.httpClient.Do(cc.tokenRequest)
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

func (cc connectorClient) GetInfo(t *testing.T, url string, headers map[string]string) (*InfoResponse, *Error) {
	request := getRequestWithHeaders(t, url, headers)

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

func getRequestWithHeaders(t *testing.T, url string, headers map[string]string) *http.Request {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)

	if headers != nil {
		for k, v := range headers {
			request.Header.Set(k, v)
		}
	}

	return request
}

func logResponse(t *testing.T, resp *http.Response) {
	dump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		t.Logf("failed to dump response, %s", err)
	} else {
		t.Logf("\n--------------------------------\n%s\n--------------------------------", dump)
	}
}
