package utils

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/sirupsen/logrus"
	"net/http"
)

type ConnectorClient interface {
	GetInfo(url string) (*InfoResponse, error)
	GetCertificate(url string, csr *x509.CertificateRequest) ([]*x509.Certificate, error)
}

type connectorClient struct {
	httpClient *http.Client
	logger     logrus.FieldLogger
}

func NewConnectorClient(skipVerify bool, logger logrus.FieldLogger) ConnectorClient {
	return &connectorClient{
		httpClient: newHttpClient(skipVerify),
		logger:     logger,
	}
}

func newHttpClient(skipVerify bool) *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipVerify},
	}
	client := &http.Client{Transport: tr}
	return client
}

func (cc *connectorClient) GetInfo(url string) (*InfoResponse, error) {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		cc.logger.Error(err)
		return nil, err
	}

	response, err := cc.httpClient.Do(request)
	if err != nil {
		cc.logger.Error(err)
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		err := parseErrorResponse(response)
		cc.logger.Errorf("GetInfo call received wrong status code: %s", err)
		return nil, err
	}

	infoResponse := &InfoResponse{}

	err = json.NewDecoder(response.Body).Decode(infoResponse)
	if err != nil {
		cc.logger.Error(err)
		return nil, err
	}

	return infoResponse, nil
}

func (cc *connectorClient) GetCertificate(url string, csr *x509.CertificateRequest) ([]*x509.Certificate, error) {
	b64CSR := base64.StdEncoding.EncodeToString(csr.Raw)

	body, err := json.Marshal(CsrRequest{Csr: b64CSR})
	if err != nil {
		cc.logger.Error(err)
		return nil, err
	}

	request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		cc.logger.Error(err)
		return nil, err
	}

	request.Close = true
	request.Header.Add("Content-Type", "application/json")

	response, err := cc.httpClient.Do(request)
	if err != nil {
		cc.logger.Error(err)
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusCreated {
		err := parseErrorResponse(response)
		cc.logger.Errorf("SignCSR call received wrong status code: %s", err)
		return nil, err
	}

	crtResponse := &CrtResponse{}

	err = json.NewDecoder(response.Body).Decode(&crtResponse)
	if err != nil {
		cc.logger.Error(err)
		return nil, err
	}

	certChainBytes, err := encodedCertChainToPemBytes(crtResponse.CRTChain)
	if err != nil {
		cc.logger.Error(err)
		return nil, err
	}

	certificateChain, err := x509.ParseCertificates(certChainBytes)
	if err != nil {
		cc.logger.Error(err)
		return nil, err
	}

	return certificateChain, nil
}

func parseErrorResponse(response *http.Response) error {
	errorResponse := ErrorResponse{}
	err := json.NewDecoder(response.Body).Decode(&errorResponse)
	if err != nil {
		return err
	}

	return errors.New(errorResponse.Error)
}
