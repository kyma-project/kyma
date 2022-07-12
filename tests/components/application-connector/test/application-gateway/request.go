package application_gateway

import (
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"net/http"
	"regexp"
	"strconv"
	"testing"
)

func executeGetRequest(t *testing.T, entry v1alpha1.Entry) int {
	t.Log("Calling", entry.CentralGatewayUrl)
	res, err := http.Get(entry.CentralGatewayUrl)
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
