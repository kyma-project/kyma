package tests

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/kyma-project/kyma/tests/application-gateway-tests/test/executor/testkit/registry"
	"github.com/kyma-project/kyma/tests/application-gateway-tests/test/executor/testkit/util"

	"github.com/kyma-project/kyma/tests/application-gateway-tests/test/executor/proxy"

	log "github.com/sirupsen/logrus"
)

func TestProxyService(t *testing.T) {

	testSuit := proxy.NewTestSuite(t)
	defer testSuit.Cleanup(t)
	testSuit.Setup(t)

	client := registry.NewAppRegistryClient(t, "http://application-registry-external-api:8081", testSuit.ApplicationName())

	t.Run("no-auth api test", func(t *testing.T) {
		apiId := client.CreateNotSecuredAPI(testSuit.GetMockServiceURL())
		log.Infoln("Created service with apiId: ", apiId)
		defer func() {
			client.CleanupService(apiId)
		}()

		log.Infoln("Labeling tests pod with denier label")
		testSuit.AddDenierLabel(t, apiId)

		log.Infoln("Calling Access Service")
		resp := testSuit.CallAccessService(t, apiId, "")
		util.RequireStatus(t, http.StatusOK, resp)

		log.Infoln("Successfully accessed application")
	})

	t.Run("basic auth api test", func(t *testing.T) {
		userName := "myUser"
		password := "mySecret"

		apiId := client.CreateBasicAuthSecuredAPI(testSuit.GetMockServiceURL()+"auth/basic/", userName, password)
		log.Infof("Created service with apiId: %s", apiId)
		defer func() {
			log.Infof("Cleaning up service %s", apiId)
			client.CleanupService(apiId)
		}()

		log.Infoln("Labeling tests pod with denier label")
		testSuit.AddDenierLabel(t, apiId)

		log.Infoln("Calling Access Service")
		resp := testSuit.CallAccessService(t, apiId, fmt.Sprintf("%s/%s", userName, password))
		util.RequireStatus(t, http.StatusOK, resp)

		log.Infoln("Successfully accessed application")
	})

}
