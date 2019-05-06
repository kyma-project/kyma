package connectorservice

import (
	"bytes"
	"crypto/tls"
	"crypto/x509/pkix"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kyma-project/kyma/components/connectivity-certs-controller/internal/certificates"

	"github.com/pkg/errors"
)

type EstablishedConnectionClient interface {
	GetManagementInfo(managementInfoURL string) (ManagementInfo, error)
	RenewCertificate(renewalURL string) (certificates.Certificates, error)
}

type establishedConnectionClient struct {
	httpClient  *http.Client
	csrProvider certificates.CSRProvider
	subject     pkix.Name
}

func NewEstablishedConnectionClient(config *tls.Config, csrProvider certificates.CSRProvider, subject pkix.Name) EstablishedConnectionClient {
	return &establishedConnectionClient{
		httpClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: config,
			},
		},
		csrProvider: csrProvider,
		subject:     subject,
	}
}

func (cc *establishedConnectionClient) GetManagementInfo(managementInfoURL string) (ManagementInfo, error) {
	request, err := http.NewRequest(http.MethodGet, managementInfoURL, nil)
	if err != nil {
		return ManagementInfo{}, errors.Wrap(err, " Failed to create Management Info request")
	}

	response, err := cc.httpClient.Do(request)
	if err != nil {
		return ManagementInfo{}, errors.Wrap(err, "Failed to request Management Info")
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

func (cc *establishedConnectionClient) RenewCertificate(renewalURL string) (certificates.Certificates, error) {
	csr, clientKey, err := cc.csrProvider.CreateCSR(cc.subject)
	if err != nil {
		return certificates.Certificates{}, errors.Wrap(err, "Failed to create CSR")
	}

	response, err := cc.requestCertificateRenewal(renewalURL, csr)
	if err != nil {
		return certificates.Certificates{}, errors.Wrap(err, "Failed to request certificate renewal")
	}

	if response.StatusCode != http.StatusCreated {
		message := fmt.Sprintf("Connector Service renewal endpoint responded with status %d", response.StatusCode)
		return certificates.Certificates{}, errors.Wrap(extractErrorResponse(response), message)
	}

	var certificateResponse CertificatesResponse
	err = readResponseBody(response.Body, &certificateResponse)
	if err != nil {
		return certificates.Certificates{}, errors.Wrap(err, "Failed to read response body")
	}

	return decodeCertificateResponse(clientKey, certificateResponse)
}

func (cc *establishedConnectionClient) requestCertificateRenewal(renewalURL, csr string) (*http.Response, error) {
	reqBody, err := json.Marshal(CertificateRequest{CSR: csr})
	if err != nil {
		return nil, errors.Wrap(err, "Failed to marshal certificate request")
	}

	request, err := http.NewRequest(http.MethodPost, renewalURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create renewal request")
	}

	return cc.httpClient.Do(request)
}
