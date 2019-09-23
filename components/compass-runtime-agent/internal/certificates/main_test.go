package certificates

import (
	"fmt"
	"os"
	"testing"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/testutil"
)

var (
	crtChain  []byte
	clientCRT []byte
	caCRT     []byte
	clientKey []byte
)

func TestMain(m *testing.M) {
	certsData, err := testutil.LoadCertsTestData("../testutil/testdata")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	crtChain = certsData.CertificateChain
	clientCRT = certsData.ClientCertificate
	caCRT = certsData.CACertificate
	clientKey = certsData.ClientKey

	exitCode := m.Run()

	os.Exit(exitCode)
}
