package application_gateway

import (
	"io"
	"net/http"
	"regexp"
	"strconv"
	"testing"

	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/pkg/errors"
)

type LogHttp struct {
	t       *testing.T
	httpCli *http.Client
}

func NewHttpCli(t *testing.T) LogHttp {
	return LogHttp{t: t, httpCli: &http.Client{}}
}

func (c LogHttp) Get(url string) (resp *http.Response, body []byte, err error) {
	c.t.Helper()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return
	}

	return c.Do(req)

}

func (c LogHttp) Do(req *http.Request) (res *http.Response, body []byte, err error) {
	c.t.Helper()
	c.t.Logf("%s %s", req.Method, req.URL)

	res, err = c.httpCli.Do(req)
	if err != nil {
		return
	}

	body, err = io.ReadAll(res.Body)
	if err == nil && len(body) > 0 {
		c.t.Logf("Body: %s", body)
	}

	return
}

func getExpectedHTTPCode(service v1alpha1.Service) (int, error) {
	re := regexp.MustCompile(`\d+`)
	if codeStr := re.FindString(service.Description); len(codeStr) > 0 {
		return strconv.Atoi(codeStr)
	}
	return 0, errors.New("Bad configuration")
}

func gatewayURL(app, service string) string {
	return "http://central-application-gateway.kyma-system:8080/" + app + "/" + service
}
