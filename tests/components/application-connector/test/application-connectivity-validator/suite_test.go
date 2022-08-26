package application_connectivity_validator

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/kyma-project/kyma/tests/components/application-connector/internal/testkit/httpd"
)

const v1EventsFormat = "http://central-application-connectivity-validator.kyma-system:8080/%s/v1/events"
const v2EventsFormat = "http://central-application-connectivity-validator.kyma-system:8080/%s/v2/events"

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

func (vs *ValidatorSuite) TestGoodCert() {
	cli := httpd.NewCli(vs.T())

	for _, testCase := range []struct {
		appName        string
		expectedHeader string
	}{{
		appName:        "event-test-standalone",
		expectedHeader: "event-test-standalone",
	}, {
		appName:        "event-test-compass",
		expectedHeader: "clientId1",
	}} {
		v1Events := fmt.Sprintf(v1EventsFormat, testCase.appName)
		v2Events := fmt.Sprintf(v2EventsFormat, testCase.appName)
		endpoints := []string{v1Events, v2Events}

		for _, url := range endpoints {
			vs.Run(fmt.Sprintf("Send request to %s URL", url), func() {
				req, err := http.NewRequest(http.MethodGet, url, nil)
				vs.Nil(err)

				req.Header.Add("X-Forwarded-Client-Cert", certFields(fmt.Sprintf("CN=%s", testCase.expectedHeader)))

				res, _, err := cli.Do(req)
				vs.Require().Nil(err)
				vs.Equal(http.StatusOK, res.StatusCode)
			})
		}
	}
}

func (vs *ValidatorSuite) TestBadCert() {
	cli := httpd.NewCli(vs.T())

	appNames := []string{"event-test-standalone", "event-test-compass"}

	for _, appName := range appNames {
		v1Events := fmt.Sprintf(v1EventsFormat, appName)
		v2Events := fmt.Sprintf(v2EventsFormat, appName)
		endpoints := []string{v1Events, v2Events}

		for _, url := range endpoints {
			vs.Run(fmt.Sprintf("Send request to %s URL with incorrect header", url), func() {
				req, err := http.NewRequest(http.MethodGet, url, nil)
				vs.Nil(err)

				req.Header.Add("X-Forwarded-Client-Cert", certFields("CN=nonexistant"))

				res, _, err := cli.Do(req)
				vs.Require().Nil(err)
				vs.Equal(http.StatusForbidden, res.StatusCode)
			})

			vs.Run(fmt.Sprintf("Send request to %s URL without header", url), func() {
				req, err := http.NewRequest(http.MethodGet, url, nil)
				vs.Nil(err)

				res, _, err := cli.Do(req)
				vs.Require().Nil(err)
				vs.Equal(http.StatusInternalServerError, res.StatusCode)
			})
		}
	}
}

func (vs *ValidatorSuite) TestInvalidPathPrefix() {
	const v3vents = "http://central-application-connectivity-validator.kyma-system:8080/event-test-compass/v3/events"

	cli := httpd.NewCli(vs.T())

	req, err := http.NewRequest(http.MethodGet, v3vents, nil)
	vs.Nil(err)

	req.Header.Add("X-Forwarded-Client-Cert", certFields("CN=clientId1"))

	res, _, err := cli.Do(req)
	vs.Require().Nil(err)
	vs.Equal(http.StatusNotFound, res.StatusCode)
}

func certFields(subject string) string {
	return fmt.Sprintf("Subject=%q", subject)
}
