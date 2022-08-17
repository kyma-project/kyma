package application_connectivity_validator

import (
	"crypto/tls"
	_ "embed"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ValidatorSuite struct {
	suite.Suite
	cli *cli.Clientset
}

func (vs *ValidatorSuite) SetupSuite() {
}

func (vs *ValidatorSuite) TearDownSuite() {
	_, err := http.Post("http://localhost:15000/quitquitquit", "", nil)
	vs.Nil(err)
	_, err = http.Post("http://localhost:15020/quitquitquit", "", nil)
	vs.Nil(err)
}

func TestValidatorSuite(t *testing.T) {
	suite.Run(t, new(ValidatorSuite))
}

//go:embed events.crt
var eventsCrt []byte

//go:embed client.key
var eventsKey []byte

func (vs *ValidatorSuite) TestValidator() {
	cert, err := tls.X509KeyPair(eventsCrt, eventsKey)

	cli := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates: []tls.Certificate{cert},
			},
		},
	}

	// const url = "localhost:8080"
	const url = "central-application-connectivity-validator.kyma-system.svc.cluster.local:8080"
	req, err := http.NewRequest(http.MethodGet, "https://"+url+"/event-test/events", nil)
	vs.Nil(err)

	req.Header.Add("X-Forwarded-Client-Cert", "Subject=\"CN=event-test\"")

	res, err := cli.Do(req)
	vs.Nil(err)
	vs.Equal(http.StatusOK, res.StatusCode)
}
