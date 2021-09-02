package testkit

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"testing"

	"github.com/avast/retry-go"
	"github.com/stretchr/testify/require"
)

const (
	ApplicationHeader = "Application"
	GroupHeader       = "Group"
	TenantHeader      = "Tenant"
	Tenant            = "testkit-tenant"
	Group             = "testkit-group"
	Extensions        = ""
	KeyAlgorithm      = "rsa2048"
)

type ConnectorClient interface {
	CreateToken(t *testing.T) (TokenResponse, error)
	GetInfo(t *testing.T, url string) (*InfoResponse, *Error)
	RevokeCertificate(t *testing.T, revocationUrl, csr string) *Error
	CreateCertChain(t *testing.T, csr, url string) (*CrtResponse, *Error)
}

type connectorClient struct {
	httpClient             *http.Client
	createTokenRequestFunc func() (*http.Request, error)
}

type createTokenRequestFunc func() (*http.Request, error)

func NewConnectorClient(createTokenRequestFunc createTokenRequestFunc, skipVerify bool) ConnectorClient {
	client := NewHttpClient(skipVerify)

	return connectorClient{
		httpClient:             client,
		createTokenRequestFunc: createTokenRequestFunc,
	}
}

func NewHttpClient(skipVerify bool) *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipVerify},
	}
	client := &http.Client{Transport: tr}
	return client
}

func (cc connectorClient) CreateToken(t *testing.T) (TokenResponse, error) {
	var response *http.Response

	err := retry.Do(func() error {
		request, err := cc.createTokenRequestFunc()
		if err != nil {
			return err
		}

		response, err = cc.httpClient.Do(request)

		return err
	})

	defer closeResponseBody(response)

	if response.StatusCode != http.StatusCreated {
		logResponse(t, response)
	}

	require.Equal(t, http.StatusCreated, response.StatusCode)

	tokenResponse := TokenResponse{}
	err = json.NewDecoder(response.Body).Decode(&tokenResponse)

	return tokenResponse, err
}

func (cc connectorClient) RevokeCertificate(t *testing.T, revocationUrl, hash string) *Error {
	body, err := json.Marshal(RevocationBody{Hash: hash})
	require.NoError(t, err)

	var response *http.Response

	err = retry.Do(func() error {
		request, err := http.NewRequest(http.MethodPost, revocationUrl, bytes.NewBuffer(body))
		if err != nil {
			return err
		}
		request.Close = true
		request.Header.Add("Content-Type", "application/json")

		response, err = cc.httpClient.Do(request)

		return err
	})

	defer closeResponseBody(response)

	require.NoError(t, err)

	if response.StatusCode != http.StatusCreated {
		return parseErrorResponse(t, response)
	}

	return nil
}

func (cc connectorClient) GetInfo(t *testing.T, url string) (*InfoResponse, *Error) {
	var response *http.Response

	err := retry.Do(func() error {
		request, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return err
		}

		request.Close = true

		response, err = cc.httpClient.Do(request)

		return err
	})

	defer closeResponseBody(response)

	require.NoError(t, err)

	if response.StatusCode != http.StatusOK {
		return nil, parseErrorResponse(t, response)
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

	var response *http.Response

	retry.Do(func() error {
		request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
		if err != nil {
			return err
		}

		request.Close = true
		request.Header.Add("Content-Type", "application/json")

		response, err = cc.httpClient.Do(request)

		return err
	})

	defer closeResponseBody(response)

	require.NoError(t, err)

	if response.StatusCode != http.StatusCreated {
		return nil, parseErrorResponse(t, response)
	}

	require.Equal(t, http.StatusCreated, response.StatusCode)

	crtResponse := &CrtResponse{}

	err = json.NewDecoder(response.Body).Decode(&crtResponse)
	require.NoError(t, err)

	return crtResponse, nil
}

func closeResponseBody(response *http.Response) {
	if response != nil {
		response.Body.Close()
	}
}

func parseErrorResponse(t *testing.T, response *http.Response) *Error {
	logResponse(t, response)
	errorResponse := ErrorResponse{}
	err := json.NewDecoder(response.Body).Decode(&errorResponse)
	require.NoError(t, err)

	return &Error{response.StatusCode, errorResponse}
}

func logResponse(t *testing.T, resp *http.Response) {
	respDump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		t.Logf("failed to dump response, %s", err)
	}

	reqDump, err := httputil.DumpRequest(resp.Request, true)
	if err != nil {
		t.Logf("failed to dump request, %s", err)
	}

	if err == nil {
		t.Logf("\n--------------------------------\n%s\n--------------------------------\n%s\n--------------------------------", reqDump, respDump)
	}
}
