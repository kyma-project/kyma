// +build integration

package integration_test

import "testing"

func TestGetClusterCatalogHappyPath(t *testing.T) {
	// given
	suite := newTestSuite(t)
	defer suite.tearDown()
	suite.AssertNoServicesInCatalogEndpoint("cluster")

	// when
	suite.createClusterAddonsConfiguration()

	// then
	suite.WaitForClusterAddonsConfigurationStatusReady()
	suite.WaitForServicesInCatalogEndpoint("cluster")

	// when
	suite.removeRepoFromClusterAddonsConfiguration("stage")

	// then
	suite.WaitForEmptyCatalogResponse("cluster")
}

// TestGetNamespacedCatalogHappyPath tests creating addons configuration in two namespaces:
// 1. create AddonsConfiguration in stage
// 2. assert services for stage
// 3. create AddonsConfiguration in prod
// 4. assert services for prod
// 5. Remove AddonsConfigurations
func TestGetNamespacedCatalogHappyPath(t *testing.T) {
	// given
	suite := newTestSuite(t)
	defer suite.tearDown()
	suite.AssertNoServicesInCatalogEndpoint("ns/stage")

	// when
	suite.createAddonsConfiguration("stage")

	// then
	suite.WaitForAddonsConfigurationStatusReady("stage")
	suite.WaitForServicesInCatalogEndpoint("ns/stage")
	suite.AssertNoServicesInCatalogEndpoint("ns/prod")

	// when
	suite.createAddonsConfiguration("prod")
	suite.WaitForAddonsConfigurationStatusReady("prod")
	suite.WaitForServicesInCatalogEndpoint("ns/prod")

	// when
	suite.removeRepoFromAddonsConfiguration("stage")
	suite.removeRepoFromAddonsConfiguration("prod")

	// then
	suite.WaitForEmptyCatalogResponse("ns/stage")
	suite.WaitForEmptyCatalogResponse("ns/prod")
}
