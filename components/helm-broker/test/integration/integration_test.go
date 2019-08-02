// +build integration

package integration_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/helm-broker/pkg/apis/addons/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func TestGetCatalogHappyPath(t *testing.T) {
	// given
	suite := newTestSuite(t)
	defer suite.tearDown()

	for name, c := range map[string]struct {
		kind      string
		addonName string
		redisID   string
		testID    string
	}{
		"namespaced-http": {
			kind:      sourceHTTP,
			addonName: addonsConfigName,
			redisID:   redisAddonID,
			testID:    accTestAddonID,
		},
		"namespaced-git": {
			kind:      sourceGit,
			addonName: addonsConfigNameGit,
			redisID:   redisAddonIDGit,
			testID:    accTestAddonIDGit,
		},
	} {
		t.Run(name, func(t *testing.T) {
			var repository *gitRepo
			if c.kind == sourceGit {
				repo, err := newGitRepository(t, addonSource)
				assert.NoError(t, err)

				defer repo.removeTmpDir()
				repository = repo
			}

			suite.assertNoServicesInCatalogEndpoint("ns/stage")

			// when
			source := newSource(c.kind, suite, repository, []string{redisAndAccTestRepo})
			suite.createAddonsConfiguration("stage", c.addonName, source)

			// then
			suite.waitForAddonsConfigurationPhase("stage", c.addonName, v1alpha1.AddonsConfigurationReady)
			suite.waitForServicesInCatalogEndpoint("ns/stage", []string{c.redisID, c.testID})
			suite.assertNoServicesInCatalogEndpoint("ns/prod")
			suite.assertNoServicesInCatalogEndpoint("cluster")

			// when
			suite.createAddonsConfiguration("prod", c.addonName, source)
			suite.waitForAddonsConfigurationPhase("prod", c.addonName, v1alpha1.AddonsConfigurationReady)
			suite.waitForServicesInCatalogEndpoint("ns/prod", []string{c.redisID, c.testID})

			// when
			source.removeURL(redisAndAccTestRepo)
			suite.updateAddonsConfigurationRepositories("stage", c.addonName, source)
			suite.updateAddonsConfigurationRepositories("prod", c.addonName, source)

			// then
			suite.waitForEmptyCatalogResponse("ns/stage")
			suite.waitForEmptyCatalogResponse("ns/prod")
		})
	}

	for name, c := range map[string]struct {
		kind      string
		addonName string
		redisID   string
		testID    string
	}{
		"cluster-http": {
			kind:      sourceHTTP,
			addonName: addonsConfigName,
			redisID:   redisAddonID,
		},
		"cluster-git": {
			kind:      sourceGit,
			addonName: addonsConfigNameGit,
			redisID:   redisAddonIDGit,
		},
	} {
		t.Run(name, func(t *testing.T) {
			var repository *gitRepo
			if c.kind == sourceGit {
				repo, err := newGitRepository(t, addonSource)
				assert.NoError(t, err)

				defer repo.removeTmpDir()
				repository = repo
			}

			suite.assertNoServicesInCatalogEndpoint("cluster")

			// when
			source := newSource(c.kind, suite, repository, []string{redisRepo})
			suite.createClusterAddonsConfiguration(c.addonName, source)

			// then
			suite.waitForClusterAddonsConfigurationPhase(c.addonName, v1alpha1.AddonsConfigurationReady)
			suite.waitForServicesInCatalogEndpoint("cluster", []string{c.redisID})

			// when
			source.removeURL(redisRepo)
			suite.updateClusterAddonsConfigurationRepositories(c.addonName, source)

			// then
			suite.waitForEmptyCatalogResponse("cluster")
		})
	}
}

func TestAddonsConflicts(t *testing.T) {
	// given
	suite := newTestSuite(t)
	defer suite.tearDown()

	for name, c := range map[string]struct {
		kind    string
		redisID string
		testID  string
	}{
		"namespaced-http": {
			kind:    sourceHTTP,
			redisID: redisAddonID,
			testID:  accTestAddonID,
		},
		"namespaced-git": {
			kind:    sourceGit,
			redisID: redisAddonIDGit,
			testID:  accTestAddonIDGit,
		},
	} {
		t.Run(name, func(t *testing.T) {
			var repository *gitRepo
			if c.kind == sourceGit {
				repo, err := newGitRepository(t, addonSource)
				assert.NoError(t, err)

				defer repo.removeTmpDir()
				repository = repo
			}
			first := "first-" + c.kind
			second := "second-" + c.kind
			third := "third-" + c.kind

			// when
			//  - create an addons configuration with repo with redis addon
			source := newSource(c.kind, suite, repository, []string{redisRepo})
			suite.createAddonsConfiguration("stage", first, source)

			// then
			//  - wait for readiness and wait for service redis at the catalog endpoint
			suite.waitForAddonsConfigurationPhase("stage", first, v1alpha1.AddonsConfigurationReady)
			suite.waitForServicesInCatalogEndpoint("ns/stage", []string{c.redisID})

			// when
			// - create second addons configuration with a repo with redis and acc-test addons
			sourceFull := newSource(c.kind, suite, repository, []string{redisAndAccTestRepo})
			suite.createAddonsConfiguration("stage", second, sourceFull)

			// then
			// - expect phase "failed", still redis service at the catalog endpoint
			suite.waitForAddonsConfigurationPhase("stage", second, v1alpha1.AddonsConfigurationFailed)
			suite.waitForServicesInCatalogEndpoint("ns/stage", []string{c.redisID})

			// when
			// - remove repo with redis from the first (cluster) addon
			source.removeURL(redisRepo)
			suite.updateAddonsConfigurationRepositories("stage", first, source)

			// then
			// - expect for readiness and 2 services at the catalog endpoint
			suite.waitForAddonsConfigurationPhase("stage", second, v1alpha1.AddonsConfigurationReady)
			suite.waitForServicesInCatalogEndpoint("ns/stage", []string{c.redisID, c.testID})

			// when
			// - create third addons configuration with a repo with acc-test addons
			sourceTesting := newSource(c.kind, suite, repository, []string{accTestRepo})
			suite.createAddonsConfiguration("stage", third, sourceTesting)

			// then
			// - expect failed (because of the conflict)
			suite.waitForAddonsConfigurationPhase("stage", third, v1alpha1.AddonsConfigurationFailed)

			// when
			// - delete second (cluster) addons configuration, so the third will be reprocessed
			suite.deleteAddonsConfiguration("stage", second)

			// then
			// - expect readiness
			suite.waitForAddonsConfigurationPhase("stage", third, v1alpha1.AddonsConfigurationReady)
			suite.waitForServicesInCatalogEndpoint("ns/stage", []string{c.testID})
		})
	}

	for name, c := range map[string]struct {
		kind    string
		redisID string
		testID  string
	}{
		"cluster-http": {
			kind:    sourceHTTP,
			redisID: redisAddonID,
			testID:  accTestAddonID,
		},
		"cluster-git": {
			kind:    sourceGit,
			redisID: redisAddonIDGit,
			testID:  accTestAddonIDGit,
		},
	} {
		t.Run(name, func(t *testing.T) {
			var repository *gitRepo
			if c.kind == sourceGit {
				repo, err := newGitRepository(t, addonSource)
				assert.NoError(t, err)

				defer repo.removeTmpDir()
				repository = repo
			}
			first := "first-" + c.kind
			second := "second-" + c.kind
			third := "third-" + c.kind

			// when
			//  - create an cluster addons configuration with repo with redis addon
			source := newSource(c.kind, suite, repository, []string{redisRepo})
			suite.createClusterAddonsConfiguration(first, source)

			// then
			//  - wait for readiness and wait for service redis at the catalog endpoint
			suite.waitForClusterAddonsConfigurationPhase(first, v1alpha1.AddonsConfigurationReady)
			suite.waitForServicesInCatalogEndpoint("cluster", []string{c.redisID})

			// when
			// - create second cluster addons configuration with a repo with redis and acc-test addons
			sourceFull := newSource(c.kind, suite, repository, []string{redisAndAccTestRepo})
			suite.createClusterAddonsConfiguration(second, sourceFull)

			// then
			// - expect phase "failed", still redis service at the catalog endpoint
			suite.waitForClusterAddonsConfigurationPhase(second, v1alpha1.AddonsConfigurationFailed)
			suite.waitForServicesInCatalogEndpoint("cluster", []string{c.redisID})

			// when
			// - remove repo with redis from the first (cluster) addon
			source.removeURL(redisRepo)
			suite.updateClusterAddonsConfigurationRepositories(first, source)

			// then
			// - expect for readiness and 2 services at the catalog endpoint
			suite.waitForClusterAddonsConfigurationPhase(second, v1alpha1.AddonsConfigurationReady)
			suite.waitForServicesInCatalogEndpoint("cluster", []string{c.redisID, c.testID})

			// when
			// - create third cluster addons configuration with a repo with acc-test addons
			sourceTesting := newSource(c.kind, suite, repository, []string{accTestRepo})
			suite.createClusterAddonsConfiguration(third, sourceTesting)

			// then
			// - expect failed (because of the conflict)
			suite.waitForClusterAddonsConfigurationPhase(third, v1alpha1.AddonsConfigurationFailed)

			// when
			// - delete second cluster addons configuration, so the third will be reprocessed
			suite.deleteClusterAddonsConfiguration(second)

			// then
			// - expect readiness
			suite.waitForClusterAddonsConfigurationPhase(third, v1alpha1.AddonsConfigurationReady)
			suite.waitForServicesInCatalogEndpoint("cluster", []string{c.testID})
		})
	}
}

func TestDocsTopic(t *testing.T) {
	// given
	suite := newTestSuite(t)
	defer suite.tearDown()

	for name, c := range map[string]struct {
		kind        string
		addonName   string
		docsTopicID string
	}{
		"namespaced-http": {
			kind:        sourceHTTP,
			addonName:   addonsConfigName,
			docsTopicID: accTestAddonID,
		},
		"namespaced-git": {
			kind:        sourceGit,
			addonName:   addonsConfigNameGit,
			docsTopicID: accTestAddonIDGit,
		},
	} {
		t.Run(name, func(t *testing.T) {
			var repository *gitRepo
			if c.kind == sourceGit {
				repo, err := newGitRepository(t, addonSource)
				assert.NoError(t, err)

				defer repo.removeTmpDir()
				repository = repo
			}

			// when
			source := newSource(c.kind, suite, repository, []string{redisAndAccTestRepo})
			suite.createAddonsConfiguration("stage", c.addonName, source)

			// then
			suite.waitForAddonsConfigurationPhase("stage", c.addonName, v1alpha1.AddonsConfigurationReady)
			suite.assertDocsTopicExist("stage", c.docsTopicID)

			// when
			source.replaceURL(redisRepo)
			suite.updateAddonsConfigurationRepositories("stage", c.addonName, source)

			// then
			suite.assertDocsTopicListIsEmpty()
		})
	}

	for name, c := range map[string]struct {
		kind        string
		addonName   string
		docsTopicID string
	}{
		"cluster-http": {
			kind:        sourceHTTP,
			addonName:   addonsConfigName,
			docsTopicID: accTestAddonID,
		},
		"cluster-git": {
			kind:        sourceGit,
			addonName:   addonsConfigNameGit,
			docsTopicID: accTestAddonIDGit,
		},
	} {
		t.Run(name, func(t *testing.T) {
			var repository *gitRepo
			if c.kind == sourceGit {
				repo, err := newGitRepository(t, addonSource)
				assert.NoError(t, err)

				defer repo.removeTmpDir()
				repository = repo
			}

			// when
			source := newSource(c.kind, suite, repository, []string{redisAndAccTestRepo})
			suite.createClusterAddonsConfiguration(c.addonName, source)

			// then
			suite.waitForClusterAddonsConfigurationPhase(c.addonName, v1alpha1.AddonsConfigurationReady)
			suite.assertClusterDocsTopicExist(c.docsTopicID)

			// when
			source.replaceURL(redisRepo)
			suite.updateClusterAddonsConfigurationRepositories(c.addonName, source)

			// then
			suite.assertClusterDocsTopicListIsEmpty()
		})
	}
}
