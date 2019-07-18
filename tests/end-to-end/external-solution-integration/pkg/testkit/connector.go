package testkit

import (
	"bytes"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"github.com/avast/retry-go"
	connectionTokenHandlerApi "github.com/kyma-project/kyma/components/connection-token-handler/pkg/apis/applicationconnector/v1alpha1"

	connectionTokenHandlerClient "github.com/kyma-project/kyma/components/connection-token-handler/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/internal/consts"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type ConnectorClient struct {
	tokenRequests connectionTokenHandlerClient.TokenRequestInterface
	httpClient    *http.Client
	logger        logrus.FieldLogger
}

func NewConnectorClient(tokenRequests connectionTokenHandlerClient.TokenRequestInterface, httpClient *http.Client, logger logrus.FieldLogger) *ConnectorClient {
	return &ConnectorClient{
		tokenRequests: tokenRequests,
		httpClient:    httpClient,
		logger:        logger,
	}
}

func (cc *ConnectorClient) GetToken() (string, error) {
	tokenRequest := &connectionTokenHandlerApi.TokenRequest{
		ObjectMeta: metav1.ObjectMeta{Name: consts.AppName},
		Status: connectionTokenHandlerApi.TokenRequestStatus{
			ExpireAfter: metav1.NewTime(time.Now().Add(1 * time.Minute)),
		},
	}

	_, err := cc.tokenRequests.Create(tokenRequest)
	if err != nil {
		cc.logger.Error(err)
		return "", err
	}

	err = retry.Do(cc.isTokenRequestReady)
	if err != nil {
		cc.logger.Error(err)
		return "", err
	}

	tr, err := cc.tokenRequests.Get(consts.AppName, metav1.GetOptions{})
	if err != nil {
		cc.logger.Error(err)
		return "", err
	}

	return tr.Status.URL, nil
}

func (cc *ConnectorClient) isTokenRequestReady() error {
	tokenRequest, e := cc.tokenRequests.Get(consts.AppName, metav1.GetOptions{})
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
