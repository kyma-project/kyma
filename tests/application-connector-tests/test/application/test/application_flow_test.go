package test

import (
	"testing"

	"github.com/kyma-project/kyma/tests/application-connector-tests/test/application"
)

func TestApplicationFlow(t *testing.T) {

	t.Logf("Stating Application flow tests")
	testSuite := application.NewTestSuite(t)
	defer testSuite.Cleanup(t)

	t.Logf("Creating test application...")
	testSuite.CreateTestApplication(t)
	t.Logf("Application created")

	t.Logf("Exchanging certificates to establish connection...")
	testSuite.EstablishMTLSConnection(t)
	t.Logf("Connection established")



}
