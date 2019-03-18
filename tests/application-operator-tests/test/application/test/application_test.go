package test

import (
	"testing"

	"github.com/kyma-project/kyma/tests/application-operator-tests/test/application"

	log "github.com/sirupsen/logrus"
)

func TestApplicationRelease(t *testing.T) {

	testSuite := application.NewTestSuite(t)
	defer testSuite.Cleanup(t)
	testSuite.Setup(t)

	t.Run("Application tests should succeed", func(t *testing.T) {
		log.Infoln("Waiting for application to be deployed...")
		testSuite.WaitForApplicationToBeDeployed(t)

		log.Infoln("Running Application test...")
		testSuite.RunApplicationTests(t)
	})
}
