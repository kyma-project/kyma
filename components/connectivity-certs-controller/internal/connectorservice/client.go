package connectorservice

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/kyma-project/kyma/components/connectivity-certs-controller/internal/certificates"

	"github.com/pkg/errors"
)

type Client interface {
	ConnectToCentralConnector(csrInfoURL string) (EstablishedConnection, error)
}

type connectorClient struct {
	csrProvider certificates.CSRProvider
	httpClient  *http.Client
}

func NewConnectorClient(csrProvider certificates.CSRProvider) Client {
	return &connectorClient{
		httpClient:  &http.Client{},
		csrProvider: csrProvider,
	}
}

func (cc *connectorClient) ConnectToCentralConnector(csrInfoURL string) (EstablishedConnection, error) {
	infoResponse, err := cc.requestCSRInfo(csrInfoURL)
	if err != nil {
		return EstablishedConnection{}, errors.Wrap(err, "Failed while requesting CSR info")
	}

	encodedCSR, err := cc.csrProvider.CreateCSR(infoResponse.CertificateInfo.Subject)
	if err != nil {
		return EstablishedConnection{}, errors.Wrap(err, "Failed while creating CSR")
	}

	certificateResponse, err := cc.requestCertificates(infoResponse.CsrURL, encodedCSR)
	if err != nil {
		return EstablishedConnection{}, errors.Wrap(err, "Failed while requesting certificate")
	}

	return composeConnectionData(certificateResponse, infoResponse.Api.InfoURL)
}

func (cc *connectorClient) requestCSRInfo(csrInfoURL string) (InfoResponse, error) {
	csrInfoReq, err := http.NewRequest(http.MethodGet, csrInfoURL, nil)
	if err != nil {
		return InfoResponse{}, err
	}

	response, err := cc.httpClient.Do(csrInfoReq)
	if err != nil {
		return InfoResponse{}, err
	}

	if response.StatusCode != http.StatusOK {
		message := fmt.Sprintf("Connector Service CSR info responded with status %d", response.StatusCode)
		return InfoResponse{}, errors.Wrap(extractErrorResponse(response), message)
	}

	var infoResponse InfoResponse
	err = readResponseBody(response.Body, &infoResponse)
	if err != nil {
		return InfoResponse{}, err
	}

	return infoResponse, nil
}

func (cc *connectorClient) requestCertificates(csrURL string, encodedCSR string) (CertificatesResponse, error) {
	requestData, err := json.Marshal(CertificateRequest{CSR: encodedCSR})
	if err != nil {
		return CertificatesResponse{}, err
	}

	certificateReq, err := http.NewRequest(http.MethodPost, csrURL, bytes.NewBuffer(requestData))
	if err != nil {
		return CertificatesResponse{}, err
	}

	response, err := cc.httpClient.Do(certificateReq)
	if err != nil {
		return CertificatesResponse{}, err
	}

	if response.StatusCode != http.StatusCreated {
		message := fmt.Sprintf("Connector Service Certificates responded with status %d", response.StatusCode)
		return CertificatesResponse{}, errors.Wrap(extractErrorResponse(response), message)
	}

	var certificateResponse CertificatesResponse
	err = readResponseBody(response.Body, &certificateResponse)
	if err != nil {
		return CertificatesResponse{}, err
	}

	return certificateResponse, nil
}

func composeConnectionData(certificateResponse CertificatesResponse, managementInfoURL string) (EstablishedConnection, error) {
	crtChain, err := base64.StdEncoding.DecodeString(certificateResponse.CRTChain)
	if err != nil {
		return EstablishedConnection{}, errors.Wrap(err, "Failed to decode certificate chain")
	}

	clientCRT, err := base64.StdEncoding.DecodeString(certificateResponse.ClientCRT)
	if err != nil {
		return EstablishedConnection{}, errors.Wrap(err, "Failed to decode client certificate")
	}

	caCRT, err := base64.StdEncoding.DecodeString(certificateResponse.CaCRT)
	if err != nil {
		return EstablishedConnection{}, errors.Wrap(err, "Failed to decode CA certificate")
	}

	certs := certificates.Certificates{
		CRTChain:  crtChain,
		ClientCRT: clientCRT,
		CaCRT:     caCRT,
	}

	return EstablishedConnection{
		Certificates:      certs,
		ManagementInfoURL: managementInfoURL,
	}, nil
}

func extractErrorResponse(response *http.Response) error {
	var errorResponse ErrorResponse
	err := readResponseBody(response.Body, &errorResponse)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal error response")
	}

	return errors.New(errorResponse.Error)
}

func readResponseBody(body io.Reader, model interface{}) error {
	rawData, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(rawData, &model)
	if err != nil {
		return err
	}

	return nil
}
