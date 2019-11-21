package compassconnection

import (
	"crypto/rsa"
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"

	gqlschema "github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"

	"kyma-project.io/compass-runtime-agent/internal/certificates"

	"kyma-project.io/compass-runtime-agent/internal/testutil"
	compassCRClientset "kyma-project.io/compass-runtime-agent/pkg/client/clientset/versioned/typed/compass/v1alpha1"

	"github.com/sirupsen/logrus"

	"github.com/pkg/errors"

	"kyma-project.io/compass-runtime-agent/pkg/apis/compass/v1alpha1"

	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

var cfg *rest.Config
var testEnv *envtest.Environment

var compassConnectionCRClient compassCRClientset.CompassConnectionInterface

var (
	crtChain    []byte
	clientCRT   []byte
	caCRT       []byte
	clientKey   *rsa.PrivateKey
	credentials certificates.Credentials
)

var (
	connectorCertResponse gqlschema.CertificationResult
)

func TestMain(m *testing.M) {
	os.Exit(runTests(m))
}

func runTests(m *testing.M) int {
	err := setupEnv()
	if err != nil {
		logrus.Errorf("Failed to setup test environment: %s", err.Error())
		os.Exit(1)
	}
	defer func() {
		err := testEnv.Stop()
		if err != nil {
			logrus.Errorf("error while deleting Compass Connection: %s", err.Error())
		}
	}()

	compassClientset, err := compassCRClientset.NewForConfig(cfg)
	if err != nil {
		logrus.Errorf("Failed to setup CompassConnection clientset: %s", err.Error())
		os.Exit(1)
	}

	compassConnectionCRClient = compassClientset.CompassConnections()

	err = setupCredentials()
	if err != nil {
		logrus.Errorf("Failed to setup credentials: %s", err.Error())
		os.Exit(1)
	}

	return m.Run()
}

func setupEnv() error {
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("testdata")},
	}

	var err error
	cfg, err = testEnv.Start()
	if err != nil {
		return errors.Wrap(err, "Failed to start test environment")
	}

	err = v1alpha1.AddToScheme(scheme.Scheme)
	if err != nil {
		return errors.Wrap(err, "Failed to add to schema")
	}

	return nil
}

func setupCredentials() error {
	certsData, err := testutil.LoadCertsTestData("../testutil/testdata")
	if err != nil {
		return errors.Wrap(err, "Failed to load certs test data")
	}

	crtChain = certsData.CertificateChain
	clientCRT = certsData.ClientCertificate
	caCRT = certsData.CACertificate
	clientKey, err = certificates.ParsePrivateKey(certsData.ClientKey)
	if err != nil {
		return errors.Wrap(err, "Failed to parse private key")
	}

	connectorCertResponse = gqlschema.CertificationResult{
		CertificateChain:  base64.StdEncoding.EncodeToString(crtChain),
		CaCertificate:     base64.StdEncoding.EncodeToString(caCRT),
		ClientCertificate: base64.StdEncoding.EncodeToString(clientCRT),
	}

	credentials, err = certificates.NewCredentials(clientKey, connectorCertResponse)
	if err != nil {
		return errors.Wrap(err, "Failed to create credentials")
	}

	return nil
}
