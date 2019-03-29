package connectorservice

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

type MutualTLSConnectorClient interface {
	GetManagementInfo(managementInfoURL string) (ManagementInfo, error)
}

type mutualTLSConnectorClient struct {
	httpClient *http.Client
}

func NewMutualTLSConnectorClient(key *rsa.PrivateKey, certificates []*x509.Certificate) MutualTLSConnectorClient {
	var rawCerts [][]byte

	for _, c := range certificates {
		rawCerts = append(rawCerts, c.Raw)
	}

	certs := []tls.Certificate{
		{
			PrivateKey:  key,
			Certificate: rawCerts,
		},
	}

	tlsConfig := &tls.Config{
		Certificates: certs,
	}

	return &mutualTLSConnectorClient{
		httpClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlsConfig,
			},
		},
	}
}

func (cc *mutualTLSConnectorClient) GetManagementInfo(managementInfoURL string) (ManagementInfo, error) {
	request, err := http.NewRequest(http.MethodGet, managementInfoURL, nil)
	if err != nil {
		return ManagementInfo{}, errors.Wrap(err, " Failed to create Management Info request.")
	}

	response, err := cc.httpClient.Do(request)
	if err != nil {
		return ManagementInfo{}, errors.Wrap(err, "Failed to request Management Info.")
	}

	if response.StatusCode != http.StatusOK {
		message := fmt.Sprintf("Connector Service Management Info responded with status %d", response.StatusCode)
		return ManagementInfo{}, errors.Wrap(extractErrorResponse(response), message)
	}

	var managementInfoResponse ManagementInfo
	err = readResponseBody(response.Body, &managementInfoResponse)
	if err != nil {
		return ManagementInfo{}, errors.Wrap(err, "Failed to read response body")
	}

	return managementInfoResponse, nil
}
