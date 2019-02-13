package testkit

import (
	"crypto/rsa"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

type SecuredConnectorClient interface {
	GetMgmInfo(t *testing.T, url string, headers map[string]string) (*ManagementInfoResponse, *Error)
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

	mgmInfoResponse := &ManagementInfoResponse{}

	err = json.NewDecoder(response.Body).Decode(&mgmInfoResponse)
	require.NoError(t, err)

	return mgmInfoResponse, nil
}
