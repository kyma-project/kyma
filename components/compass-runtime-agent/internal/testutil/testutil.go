package testutil

import (
	"io/ioutil"
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
	crtChain, err := ioutil.ReadFile(path.Join(testDataPath, "cert.chain.pem"))
	if err != nil {
		return CertsTestData{}, errors.Errorf("Failed to read certificate chain testdata")
	}

	caCRT, err := ioutil.ReadFile(path.Join(testDataPath, "ca.crt.pem"))
	if err != nil {
		return CertsTestData{}, errors.Errorf("Failed to read CA certificate testdata")
	}

	clientCRT, err := ioutil.ReadFile(path.Join(testDataPath, "client.crt.pem"))
	if err != nil {
		return CertsTestData{}, errors.Errorf("Failed to read client certificate testdata")
	}

	clientKey, err := ioutil.ReadFile(path.Join(testDataPath, "client.key.pem"))
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
