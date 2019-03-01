package testkit

import (
	"bytes"
	"crypto/rsa"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

type SecuredConnectorClient interface {
	GetMgmInfo(t *testing.T, url string, headers map[string]string) (*ManagementInfoResponse, *Error)
	RenewCertificate(t *testing.T, url string, csr string) (*CrtResponse, *Error)
	RevokeCertificate(t *testing.T, url string, csr string) *http.Response
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

func (cc securedConnectorClient) GetMgmInfo(t *testing.T, url string, headers map[string]string) (*ManagementInfoResponse, *Error) {
	request := getRequestWithHeaders(t, url, headers)

	var mgmInfoResponse ManagementInfoResponse
	errorResp := cc.secureConnectorRequest(t, request, &mgmInfoResponse, http.StatusOK)

	return &mgmInfoResponse, errorResp
}

func (cc securedConnectorClient) RenewCertificate(t *testing.T, url string, csr string) (*CrtResponse, *Error) {
	body, err := json.Marshal(CsrRequest{Csr: csr})
	require.NoError(t, err)

	request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	require.NoError(t, err)

	var certificateResponse CrtResponse
	errorResp := cc.secureConnectorRequest(t, request, &certificateResponse, http.StatusCreated)

	return &certificateResponse, errorResp
}

func (cc securedConnectorClient) RevokeCertificate(t *testing.T, url string, csr string) *http.Response {
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer([]byte{}))
	require.NoError(t, err)
	request.Header.Add("Certificate-Header", csr)

	response, err := cc.httpClient.Do(request)
	require.NoError(t, err)

	return response
}

func (cc securedConnectorClient) secureConnectorRequest(t *testing.T, request *http.Request, data interface{}, expectedStatus int) *Error {
	response, err := cc.httpClient.Do(request)
	require.NoError(t, err)
	if response.StatusCode != expectedStatus {
		return parseErrorResponse(t, response)
	}

	err = json.NewDecoder(response.Body).Decode(&data)
	require.NoError(t, err)

	return nil
}
