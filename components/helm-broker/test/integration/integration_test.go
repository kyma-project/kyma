// +build integration

package integration_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/helm-broker/pkg/apis/addons/v1alpha1"
)

func TestGetCatalogHappyPath(t *testing.T) {
	// given
	suite := newTestSuite(t)
	defer suite.tearDown()

	t.Run("namespaced", func(t *testing.T) {
		suite.assertNoServicesInCatalogEndpoint("ns/stage")

		// when
		suite.createAddonsConfiguration("stage", addonsConfigName, []string{redisAndAccTestRepo})

		// then
		suite.waitForAddonsConfigurationPhase("stage", addonsConfigName, v1alpha1.AddonsConfigurationReady)
		suite.waitForServicesInCatalogEndpoint("ns/stage", []string{redisAddonID, accTestAddonID})
		suite.assertNoServicesInCatalogEndpoint("ns/prod")
		suite.assertNoServicesInCatalogEndpoint("cluster")

		// when
		suite.createAddonsConfiguration("prod", addonsConfigName, []string{redisAndAccTestRepo})
		suite.waitForAddonsConfigurationPhase("prod", addonsConfigName, v1alpha1.AddonsConfigurationReady)
		suite.waitForServicesInCatalogEndpoint("ns/prod", []string{redisAddonID, accTestAddonID})

		// when
		suite.removeRepoFromAddonsConfiguration("stage", addonsConfigName, redisAndAccTestRepo)
		suite.removeRepoFromAddonsConfiguration("prod", addonsConfigName, redisAndAccTestRepo)

		// then
		suite.waitForEmptyCatalogResponse("ns/stage")
		suite.waitForEmptyCatalogResponse("ns/prod")
	})

	t.Run("cluster", func(t *testing.T) {
		suite.assertNoServicesInCatalogEndpoint("cluster")

		// when
		suite.createClusterAddonsConfiguration(addonsConfigName, []string{redisRepo})

		// then
		suite.waitForClusterAddonsConfigurationPhase(addonsConfigName, v1alpha1.AddonsConfigurationReady)
		suite.waitForServicesInCatalogEndpoint("cluster", []string{redisAddonID})

		// when
		suite.removeRepoFromClusterAddonsConfiguration(addonsConfigName, redisRepo)

		// then
		suite.waitForEmptyCatalogResponse("cluster")
	})
}

func TestAddonsConflicts(t *testing.T) {
	// given
	suite := newTestSuite(t)
	defer suite.tearDown()

	t.Run("namespaced", func(t *testing.T) {
		// when
		//  - create an addons configuration with repo with redis addon
		suite.createAddonsConfiguration("stage", "first", []string{redisRepo})

		// then
		//  - wait for readiness and wait for service redis at the catalog endpoint
		suite.waitForAddonsConfigurationPhase("stage", "first", v1alpha1.AddonsConfigurationReady)
		suite.waitForServicesInCatalogEndpoint("ns/stage", []string{redisAddonID})

		// when
		// - create second addons configuration with a repo with redis and acc-test addons
		suite.createAddonsConfiguration("stage", "second", []string{redisAndAccTestRepo})

		// then
		// - expect phase "failed", still redis service at the catalog endpoint
		suite.waitForAddonsConfigurationPhase("stage", "second", v1alpha1.AddonsConfigurationFailed)
		suite.waitForServicesInCatalogEndpoint("ns/stage", []string{redisAddonID})

		// when
		// - remove repo with redis from the first (cluster) addon
		suite.removeRepoFromAddonsConfiguration("stage", "first", redisRepo)

		// then
		// - expect for readiness and 2 services at the catalog endpoint
		suite.waitForAddonsConfigurationPhase("stage", "second", v1alpha1.AddonsConfigurationReady)
		suite.waitForServicesInCatalogEndpoint("ns/stage", []string{redisAddonID, accTestAddonID})

		// when
		// - create third addons configuration with a repo with acc-test addons
		suite.createAddonsConfiguration("stage", "third", []string{accTestRepo})

		// then
		// - expect failed (because of the conflict)
		suite.waitForAddonsConfigurationPhase("stage", "third", v1alpha1.AddonsConfigurationFailed)

		// when
		// - delete second (cluster) addons configuration, so the third will be reprocessed
		suite.deleteAddonsConfiguration("stage", "second")

		// then
		// - expect readiness
		suite.waitForAddonsConfigurationPhase("stage", "third", v1alpha1.AddonsConfigurationReady)
		suite.waitForServicesInCatalogEndpoint("ns/stage", []string{accTestAddonID})
	})

	t.Run("cluster", func(t *testing.T) {
		// when
		//  - create an cluster addons configuration with repo with redis addon
		suite.createClusterAddonsConfiguration("first", []string{redisRepo})

		// then
		//  - wait for readiness and wait for service redis at the catalog endpoint
		suite.waitForClusterAddonsConfigurationPhase("first", v1alpha1.AddonsConfigurationReady)
		suite.waitForServicesInCatalogEndpoint("cluster", []string{redisAddonID})

		// when
		// - create second cluster addons configuration with a repo with redis and acc-test addons
		suite.createClusterAddonsConfiguration("second", []string{redisAndAccTestRepo})

		// then
		// - expect phase "failed", still redis service at the catalog endpoint
		suite.waitForClusterAddonsConfigurationPhase("second", v1alpha1.AddonsConfigurationFailed)
		suite.waitForServicesInCatalogEndpoint("cluster", []string{redisAddonID})

		// when
		// - remove repo with redis from the first (cluster) addon
		suite.removeRepoFromClusterAddonsConfiguration("first", redisRepo)

		// then
		// - expect for readiness and 2 services at the catalog endpoint
		suite.waitForClusterAddonsConfigurationPhase("second", v1alpha1.AddonsConfigurationReady)
		suite.waitForServicesInCatalogEndpoint("cluster", []string{redisAddonID, accTestAddonID})

		// when
		// - create third cluster addons configuration with a repo with acc-test addons
		suite.createClusterAddonsConfiguration("third", []string{accTestRepo})

		// then
		// - expect failed (because of the conflict)
		suite.waitForClusterAddonsConfigurationPhase("third", v1alpha1.AddonsConfigurationFailed)

		// when
		// - delete second cluster addons configuration, so the third will be reprocessed
		suite.deleteClusterAddonsConfiguration("second")

		// then
		// - expect readiness
		suite.waitForClusterAddonsConfigurationPhase("third", v1alpha1.AddonsConfigurationReady)
		suite.waitForServicesInCatalogEndpoint("cluster", []string{accTestAddonID})
	})
}
