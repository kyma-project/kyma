package application_gateway

import (
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"testing"

	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func logBody(t *testing.T, body io.Reader) {
	buf, err := ioutil.ReadAll(body)
	if err == nil && len(buf) > 0 {
		t.Log("Body:", string(buf))
	}
}

func executeGetRequest(t *testing.T, entry v1alpha1.Entry) int {
	t.Log("Calling", entry.CentralGatewayUrl)
	res, err := http.Get(entry.CentralGatewayUrl)

	if err == nil {
		logBody(t, res.Body)
	}
	assert.Nil(t, err)
	return res.StatusCode
}

func getExpectedHTTPCode(service v1alpha1.Service) (int, error) {
	re := regexp.MustCompile(`\d+`)
	if codeStr := re.FindString(service.Description); len(codeStr) > 0 {
		return strconv.Atoi(codeStr)
	}
	return 0, errors.New("Bad configuration")
}

func gatewayUrl(app, service string) string {
	return "http://central-application-gateway.kyma-system:8080/" + app + "/" + service
}
