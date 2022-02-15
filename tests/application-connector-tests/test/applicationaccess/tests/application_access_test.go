package tests

import (
	"net/http"
	"testing"

	"github.com/kyma-project/kyma/tests/application-connector-tests/test/applicationaccess"
)

const (
	applicationNamePrefix = "app-conn-tests"
)

func TestApplicationAccess(t *testing.T) {

	t.Logf("Stating Application flow tests")
	testSuite := applicationaccess.NewTestSuite(t)

	application := testSuite.PrepareTestApplication(t, applicationNamePrefix)
	t.Logf("Creating test application...")
	application = testSuite.DeployApplication(t, application)
	defer testSuite.CleanupApplication(t, application.Name)
	testSuite.WaitForApplicationToBeDeployed(t, application.Name)
	t.Logf("Application created")

	t.Logf("Exchanging certificates to establish connection for %s application...", application.Name)
	applicationConnection := testSuite.EstablishMTLSConnection(t, application)
	t.Logf("Connection established")

	t.Logf("Accessing Application")
	testSuite.ShouldAccessApplication(t, applicationConnection.Credentials, applicationConnection.ManagementURLs)

	t.Run("should receive 403 if certificate issued for different Application", func(t *testing.T) {
		notDeployedApplication := testSuite.PrepareTestApplication(t, applicationNamePrefix)

		newAppConnection := testSuite.EstablishMTLSConnection(t, notDeployedApplication)

		testSuite.ShouldFailToAccessApplication(t, newAppConnection.Credentials, applicationConnection.ManagementURLs, http.StatusForbidden)
	})
}
