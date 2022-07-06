package helper

import (
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"testing"
)

func codeFromGateway(t *testing.T, entry v1alpha1.Entry) int {
	t.Log("Calling", entry.CentralGatewayUrl)
	res, err := http.Get(entry.CentralGatewayUrl)
	assert.Nil(t, err)
	body, err := ioutil.ReadAll(res.Body)
	if err == nil && len(body) > 0 {
		t.Log("Response", string(body))
	}
	return res.StatusCode
}

func codeFromService(service v1alpha1.Service) (code int, err error) {
	code = 0
	re := regexp.MustCompile(`\d+$`)
	if codeStr := re.FindString(service.Description); len(codeStr) > 0 {
		code, err = strconv.Atoi(codeStr)
		if err != nil {
			return
		}
	}
	return
}

func GetCodes(t *testing.T, entry v1alpha1.Entry, service v1alpha1.Service) (actualCode, expectedCode int) {
	expectedCode, err := codeFromService(service)
	if err != nil {
		t.Log("Error during getting the error code from description -> applicationCRD")
		t.Fail()
	}
	actualCode = codeFromGateway(t, entry)
	return
}
