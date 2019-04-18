package testsuite

import (
	"crypto/x509"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/resourceskit"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/testkit"
	"k8s.io/client-go/rest"
	"time"
)

type TestSuite interface {
	CreateApplication() error
	FetchCertificate() ([]*x509.Certificate, error)
	//RegisterService()
	//StartTestServer()
	//SendEvent()
	//CleanUp()
}

type testSuite struct {
	acClient   resourceskit.AppConnectorClient
	k8sClient  resourceskit.K8sResourcesClient
	trClient   resourceskit.TokenRequestClient
	connClient testkit.ConnectorClient
}

func NewTestSuite(config *rest.Config) TestSuite {
	return &testSuite{}
}

func (ts *testSuite) CreateApplication() error {
	_, err := ts.acClient.CreateDummyApplication(appName, accessLabel, false)
	if err != nil {
		return err
	}

	//TODO: Polling / retries
	time.Sleep(5 * time.Second)
	checker := resourceskit.NewK8sChecker(ts.k8sClient, appName)

	err = checker.CheckK8sResources()
	if err != nil {
		return err
	}

	return nil
}

func (ts *testSuite) FetchCertificate() ([]*x509.Certificate, error) {
	key, err := testkit.CreateKey()
	if err != nil {
		return nil, err
	}

	infoURL, err := ts.connClient.GetToken(appName)
	if err != nil {
		return nil, err
	}

	info, err := ts.connClient.GetInfo(infoURL)
	if err != nil {
		return nil, err
	}

	csr, err := testkit.CreateCSR(info.Certificate.Subject, key)
	if err != nil {
		return nil, err
	}

	certificate, err := ts.connClient.GetCertificate(info.CertUrl, csr)
	if err != nil {
		return nil, err
	}

	return certificate, nil
}
