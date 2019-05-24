package test

import (
	"net/http"
	"testing"

	"github.com/kyma-project/kyma/tests/application-connector-tests/test/applicationflow"
)

const (
	applicationNamePrefix = "app-conn-tests"
)

func TestApplicationFlow(t *testing.T) {

	t.Logf("Stating Application flow tests")
	testSuite := applicationflow.NewTestSuite(t)

	application := testSuite.PrepareTestApplication(t, applicationNamePrefix)
	t.Logf("Creating test application...")
	application = testSuite.DeployApplication(t, application)
	defer testSuite.CleanupApplication(t, application.Name)
	testSuite.WaitForApplicationToBeDeployed(t, application.Name)
	t.Logf("Application created")

	t.Run("should successfully access Application", func(t *testing.T) {
		t.Logf("Exchanging certificates to establish connection for %s application...", application.Name)
		appConnection := testSuite.EstablishMTLSConnection(t, application)
		t.Logf("Connection established")

		testSuite.ShouldAccessApplication(t, appConnection)
	})

	t.Run("should receive 403 if certificate issued for different Application", func(t *testing.T) {
		notDeployedApplication := testSuite.PrepareTestApplication(t, applicationNamePrefix)

		t.Logf("Exchanging certificates to establish connection for %s application...", notDeployedApplication.Name)
		appConnection := testSuite.EstablishMTLSConnection(t, notDeployedApplication)
		t.Logf("Connection established")

		testSuite.ShouldFailToAccessApplication(t, appConnection, http.StatusForbidden)
	})

	// TODO - if central check group and tenant not matching
}
