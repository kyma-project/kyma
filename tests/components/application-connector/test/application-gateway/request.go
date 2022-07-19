package application_gateway

import (
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"testing"

	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func executeGetRequest(t *testing.T, entry v1alpha1.Entry) (int, *http.Response) {
	t.Log("Calling", entry.CentralGatewayUrl)
	res, err := http.Get(entry.CentralGatewayUrl)

	if err == nil {
		body, err := ioutil.ReadAll(res.Body)
		if err == nil && len(body) > 0 {
			t.Log("Response", string(body))
		}
	}
	assert.Nil(t, err)
	return res.StatusCode, res
}

func getExpectedHTTPCode(service v1alpha1.Service) (int, error) {
	re := regexp.MustCompile(`\d+`)
	if codeStr := re.FindString(service.Description); len(codeStr) > 0 {
		return strconv.Atoi(codeStr)
	}
	return 0, errors.New("Bad configuration")
}
