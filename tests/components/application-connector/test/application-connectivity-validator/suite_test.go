package application_connectivity_validator

import (
	"fmt"
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

const url = "http://central-application-connectivity-validator.kyma-system.svc.cluster.local:8080/event-test/events"

func (vs *ValidatorSuite) TestGoodCert() {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	vs.Nil(err)

	req.Header.Add("X-Forwarded-Client-Cert", certFields("CN=event-test"))

	res, err := http.DefaultClient.Do(req)
	vs.Nil(err)
	vs.Equal(http.StatusOK, res.StatusCode)
}

func (vs *ValidatorSuite) TestBadCert() {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	vs.Nil(err)

	req.Header.Add("X-Forwarded-Client-Cert", certFields("CN=nonexistant"))

	res, err := http.DefaultClient.Do(req)
	vs.Nil(err)
	vs.Equal(http.StatusForbidden, res.StatusCode)
}

func certFields(subject string) string {
	return fmt.Sprintf("Subject=%q", subject)
}
