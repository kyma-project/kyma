package testkit

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"github.com/avast/retry-go"
	"github.com/pkg/errors"
	"net/http"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/resourceskit"
	"github.com/sirupsen/logrus"
)

type ConnectorClient struct {
	TokenRequestClient resourceskit.TokenRequestClient
	httpClient         *http.Client
	logger             logrus.FieldLogger
}

func NewConnectorClient(trClient resourceskit.TokenRequestClient, skipVerify bool, logger logrus.FieldLogger) *ConnectorClient {
	return &ConnectorClient{
		TokenRequestClient: trClient,
		httpClient:         newHttpClient(skipVerify),
		logger:             logger,
	}
}

func newHttpClient(skipVerify bool) *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipVerify},
	}
	client := &http.Client{Transport: tr}
	return client
}

func (cc *ConnectorClient) GetToken() (string, error) {
	_, err := cc.TokenRequestClient.CreateTokenRequest()
	if err != nil {
		cc.logger.Error(err)
		return "", err
	}

	err = retry.Do(cc.isTokenRequestReady)
	if err != nil {
		cc.logger.Error(err)
		return "", err
	}

	tr, err := cc.TokenRequestClient.GetTokenRequest()
	if err != nil {
		cc.logger.Error(err)
		return "", err
	}

	return tr.Status.URL, nil
}

func (cc *ConnectorClient) isTokenRequestReady() error {
	tokenRequest, e := cc.TokenRequestClient.GetTokenRequest()
	if e != nil {
		return e
	}

	if &tokenRequest.Status == nil || &tokenRequest.Status.URL == nil || tokenRequest.Status.URL == "" {
		return errors.New("token not ready yet")
	}

	return nil
}

func (cc *ConnectorClient) GetInfo(url string) (*InfoResponse, error) {
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

func (cc *ConnectorClient) GetCertificate(url string, csr []byte) ([]*x509.Certificate, error) {
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
