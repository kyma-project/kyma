package testkit

import (
	"bytes"
	"crypto/rsa"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/avast/retry-go"

	"github.com/stretchr/testify/require"
)

type SecuredConnectorClient interface {
	GetMgmInfo(t *testing.T, url string) (*ManagementInfoResponse, *Error)
	RenewCertificate(t *testing.T, url string, csr string) (*CrtResponse, *Error)
	RevokeCertificate(t *testing.T, url string) *Error
}

type securedConnectorClient struct {
	httpClient *http.Client
}

func NewSecuredConnectorClient(skipVerify bool, key *rsa.PrivateKey, certs []byte) SecuredConnectorClient {

	client := NewTLSClientWithCert(skipVerify, key, certs)

	return &securedConnectorClient{
		httpClient: client,
	}
}

func NewTLSClientWithCert(skipVerify bool, key *rsa.PrivateKey, certificate ...[]byte) *http.Client {
	tlsCert := tls.Certificate{
		Certificate: certificate,
		PrivateKey:  key,
	}

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{tlsCert},
		ClientAuth:         tls.RequireAndVerifyClientCert,
		InsecureSkipVerify: skipVerify,
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	return &http.Client{
		Transport: transport,
	}
}

func (cc securedConnectorClient) GetMgmInfo(t *testing.T, url string) (*ManagementInfoResponse, *Error) {
	createRequestFunction := func() (*http.Request, error) {
		request, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}

		request.Close = true

		return request, nil
	}

	var mgmInfoResponse ManagementInfoResponse
	errorResp := cc.secureConnectorRequest(t, createRequestFunction, &mgmInfoResponse, http.StatusOK)

	return &mgmInfoResponse, errorResp
}

func (cc securedConnectorClient) RenewCertificate(t *testing.T, url string, csr string) (*CrtResponse, *Error) {
	body, err := json.Marshal(CsrRequest{Csr: csr})
	require.NoError(t, err)

	createRequestFunction := func() (*http.Request, error) {
		request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
		if err != nil {
			return nil, err
		}
		request.Close = true
		request.Header.Add("Content-Type", "application/json")

		return request, err
	}

	var certificateResponse CrtResponse
	errorResp := cc.secureConnectorRequest(t, createRequestFunction, &certificateResponse, http.StatusCreated)

	return &certificateResponse, errorResp
}

func (cc securedConnectorClient) RevokeCertificate(t *testing.T, url string) *Error {
	createRequestFunction := func() (*http.Request, error) {
		request, err := http.NewRequest(http.MethodPost, url, nil)
		if err != nil {
			return nil, err
		}
		request.Close = true

		return request, err
	}

	errorResponse := cc.secureConnectorRequest(t, createRequestFunction, nil, http.StatusCreated)

	return errorResponse
}

func (cc securedConnectorClient) secureConnectorRequest(t *testing.T, createTokenRequestFunc createTokenRequestFunc, data interface{}, expectedStatus int) *Error {

	var response *http.Response

	err := retry.Do(func() error {
		request, err := createTokenRequestFunc()
		if err != nil {
			return err
		}
		response, err = cc.httpClient.Do(request)

		return err
	})

	defer closeResponseBody(response)

	require.NoError(t, err)

	if response.StatusCode != expectedStatus {
		return parseErrorResponse(t, response)
	}

	if data != nil {
		err = json.NewDecoder(response.Body).Decode(&data)
		require.NoError(t, err)
	}

	return nil
}
