package certificates

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/pkg/errors"
)

var (
	crtChain  []byte
	clientCRT []byte
	caCRT     []byte
	clientKey []byte
)

func TestMain(m *testing.M) {
	err := loadTestData()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	exitCode := m.Run()

	os.Exit(exitCode)
}

func loadTestData() error {
	var err error

	crtChain, err = ioutil.ReadFile("testdata/cert.chain.pem")
	if err != nil {
		return errors.Errorf("Failed to read certificate chain testdata")
	}

	caCRT, err = ioutil.ReadFile("testdata/ca.crt.pem")
	if err != nil {
		return errors.Errorf("Failed to read CA certificate testdata")
	}

	clientCRT, err = ioutil.ReadFile("testdata/client.crt.pem")
	if err != nil {
		return errors.Errorf("Failed to read client certificate testdata")
	}

	clientKey, err = ioutil.ReadFile("testdata/client.key.pem")
	if err != nil {
		return errors.Errorf("Failed to read client key testdata")
	}

	return nil
}
