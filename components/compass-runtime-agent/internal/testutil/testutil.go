package testutil

import (
	"os"
	"path"

	"github.com/pkg/errors"
)

type CertsTestData struct {
	CertificateChain  []byte
	CACertificate     []byte
	ClientCertificate []byte
	ClientKey         []byte
}

func LoadCertsTestData(testDataPath string) (CertsTestData, error) {
	crtChain, err := os.ReadFile(path.Join(testDataPath, "cert.chain.pem"))
	if err != nil {
		return CertsTestData{}, errors.Errorf("Failed to read certificate chain testdata")
	}

	caCRT, err := os.ReadFile(path.Join(testDataPath, "ca.crt.pem"))
	if err != nil {
		return CertsTestData{}, errors.Errorf("Failed to read CA certificate testdata")
	}

	clientCRT, err := os.ReadFile(path.Join(testDataPath, "client.crt.pem"))
	if err != nil {
		return CertsTestData{}, errors.Errorf("Failed to read client certificate testdata")
	}

	clientKey, err := os.ReadFile(path.Join(testDataPath, "client.key.pem"))
	if err != nil {
		return CertsTestData{}, errors.Errorf("Failed to read client key testdata")
	}

	return CertsTestData{
		CertificateChain:  crtChain,
		CACertificate:     caCRT,
		ClientCertificate: clientCRT,
		ClientKey:         clientKey,
	}, nil
}
