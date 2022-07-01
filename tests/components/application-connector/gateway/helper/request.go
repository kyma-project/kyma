package helper

import (
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
)

func CallToGateway(t *testing.T, entry v1alpha1.Entry) int {
	t.Log("Calling", entry.CentralGatewayUrl)
	res, err := http.Get(entry.CentralGatewayUrl)
	assert.Nil(t, err)
	body, err := ioutil.ReadAll(res.Body)
	if err == nil && len(body) > 0 {
		t.Log("Response", string(body))
	}
	return res.StatusCode
}
