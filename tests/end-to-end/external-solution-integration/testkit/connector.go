package testkit

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/resourceskit"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/wait"
	"github.com/sirupsen/logrus"
)

type ConnectorClient interface {
	GetToken() (string, error)
	GetInfo(url string) (*InfoResponse, error)
	GetCertificate(url string, csr []byte) ([]*x509.Certificate, error)
}

type connectorClient struct {
	trClient   resourceskit.TokenRequestClient
	httpClient *http.Client
	logger     logrus.FieldLogger
}

func NewConnectorClient(trClient resourceskit.TokenRequestClient, skipVerify bool, logger logrus.FieldLogger) ConnectorClient {
	return &connectorClient{
		trClient:   trClient,
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

func (cc *connectorClient) GetToken() (string, error) {
	_, err := cc.trClient.CreateTokenRequest()
	if err != nil {
		cc.logger.Error(err)
		return "", err
	}

	err = wait.Until(5, 10, cc.isTokenRequestReady)
	if err != nil {
		cc.logger.Error(err)
		return "", err
	}

	tr, err := cc.trClient.GetTokenRequest()
	if err != nil {
		cc.logger.Error(err)
		return "", err
	}

	return tr.Status.URL, nil
}

func (cc *connectorClient) isTokenRequestReady() (bool, error) {
	tokenRequest, e := cc.trClient.GetTokenRequest()
	if e != nil {
		return false, e
	}

	return &tokenRequest.Status != nil && &tokenRequest.Status.URL != nil && tokenRequest.Status.URL != "", nil
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
		cc.logger.Error(err)
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

func (cc *connectorClient) GetCertificate(url string, csr []byte) ([]*x509.Certificate, error) {
	b64CSR := base64.StdEncoding.EncodeToString(csr)

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

	request.Header.Add("Content-Type", "application/json")

	response, err := cc.httpClient.Do(request)
	if err != nil {
		cc.logger.Error(err)
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusCreated {
		err := parseErrorResponse(response)
		cc.logger.Error(err)
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
