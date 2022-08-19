package application_connectivity_validator

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/kyma-project/kyma/tests/components/application-connector/internal/testkit/httpd"
)

type ValidatorSuite struct {
	suite.Suite
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
	cli := httpd.NewCli(vs.T())

	req, err := http.NewRequest(http.MethodGet, url, nil)
	vs.Nil(err)

	req.Header.Add("X-Forwarded-Client-Cert", certFields("CN=event-test"))

	res, _, err := cli.Do(req)
	vs.Require().Nil(err)
	vs.Equal(http.StatusOK, res.StatusCode)
}

func (vs *ValidatorSuite) TestBadCert() {
	cli := httpd.NewCli(vs.T())

	req, err := http.NewRequest(http.MethodGet, url, nil)
	vs.Nil(err)

	req.Header.Add("X-Forwarded-Client-Cert", certFields("CN=nonexistant"))

	res, _, err := cli.Do(req)
	vs.Require().Nil(err)
	vs.Equal(http.StatusForbidden, res.StatusCode)
}

func certFields(subject string) string {
	return fmt.Sprintf("Subject=%q", subject)
}
