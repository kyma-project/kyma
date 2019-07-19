package addons

import (
	"fmt"

	addonsv1alpha1 "github.com/kyma-project/kyma/components/helm-broker/pkg/apis/addons/v1alpha1"
)

// RepositoryCollection keeps and process collection of RepositoryController
type RepositoryCollection struct {
	Repositories []*RepositoryController
}

// NewRepositoryCollection returns pointer to RepositoryCollection
func NewRepositoryCollection() *RepositoryCollection {
	return &RepositoryCollection{
		Repositories: []*RepositoryController{},
	}
}

// AddRepository adds new RepositoryController to RepositoryCollection
func (rc *RepositoryCollection) AddRepository(repo *RepositoryController) {
	rc.Repositories = append(rc.Repositories, repo)
}

func (rc *RepositoryCollection) addons() []*AddonController {
	addons := []*AddonController{}

	for _, repo := range rc.Repositories {
		for _, addon := range repo.Addons {
			addons = append(addons, addon)
		}
	}

	return addons
}

func (rc *RepositoryCollection) completeAddons() []*AddonController {
	addons := []*AddonController{}

	for _, addon := range rc.addons() {
		if !addon.IsComplete() {
			continue
		}
		addons = append(addons, addon)
	}

	return addons
}

// ReadyAddons returns all addons from all repositories which ready status
func (rc *RepositoryCollection) ReadyAddons() []*AddonController {
	addons := []*AddonController{}

	for _, addon := range rc.addons() {
		if !addon.IsReady() {
			continue
		}
		addons = append(addons, addon)
	}

	return addons
}

// IsRepositoriesFailed informs if any of repositories in collection is in failed status
func (rc *RepositoryCollection) IsRepositoriesFailed() bool {
	for _, repository := range rc.Repositories {
		if repository.HasFailedAddons() {
			repository.Failed()
			return true
		}
	}

	return false
}

type idConflictData struct {
	repositoryURL string
	addonsName    string
}

// ReviseBundleDuplicationInRepository checks all completed addons (addons without fetch/load error)
// they have no ID conflict with other addons in other or the same repository
func (rc *RepositoryCollection) ReviseBundleDuplicationInRepository() {
	ids := make(map[string]idConflictData)

	for _, addon := range rc.completeAddons() {
		if data, ok := ids[addon.ID]; ok {
			addon.ConflictInSpecifiedRepositories(fmt.Errorf("[url: %s, addons: %s]", data.repositoryURL, data.addonsName))
		} else {
			ids[addon.ID] = idConflictData{
				repositoryURL: addon.URL,
				addonsName:    fmt.Sprintf("%s:%s", addon.Addon.Name, addon.Addon.Version),
			}
		}
	}
}

// ReviseBundleDuplicationInStorage checks all completed addons (addons without fetch/load error)
// they have no name:version conflict with other AddonConfiguration
func (rc *RepositoryCollection) ReviseBundleDuplicationInStorage(acList *addonsv1alpha1.AddonsConfigurationList) {
	for _, addon := range rc.completeAddons() {
		rc.findExistingAddon(addon, acList)
	}
}

// ReviseBundleDuplicationInClusterStorage checks all completed addons (addons without fetch/load error)
// they have no name:version conflict with other AddonConfiguration
func (rc *RepositoryCollection) ReviseBundleDuplicationInClusterStorage(acList *addonsv1alpha1.ClusterAddonsConfigurationList) {
	for _, addon := range rc.completeAddons() {
		rc.findExistingClusterAddon(addon, acList)
	}
}

func (rc *RepositoryCollection) findExistingAddon(addon *AddonController, list *addonsv1alpha1.AddonsConfigurationList) {
	for _, existAddonConfiguration := range list.Items {
		for _, repo := range existAddonConfiguration.Status.Repositories {
			if rc.addonAlreadyRegistered(*addon, rc.filterReadyAddons(repo)) {
				addon.ConflictWithAlreadyRegisteredAddons(fmt.Errorf("[ConfigurationName: %s, url: %s, addons: %s:%s]", existAddonConfiguration.Name, repo.URL, addon.Addon.Name, addon.Addon.Version))
			}
		}
	}
}

func (rc *RepositoryCollection) findExistingClusterAddon(addon *AddonController, list *addonsv1alpha1.ClusterAddonsConfigurationList) {
	for _, existAddonConfiguration := range list.Items {
		for _, repo := range existAddonConfiguration.Status.Repositories {
			if rc.addonAlreadyRegistered(*addon, rc.filterReadyAddons(repo)) {
				addon.ConflictWithAlreadyRegisteredAddons(fmt.Errorf("[ConfigurationName: %s, url: %s, addons: %s:%s]", existAddonConfiguration.Name, repo.URL, addon.Addon.Name, addon.Addon.Version))
			}
		}
	}
}

func (rc *RepositoryCollection) filterReadyAddons(repository addonsv1alpha1.StatusRepository) []addonsv1alpha1.Addon {
	addons := []addonsv1alpha1.Addon{}

	for _, add := range repository.Addons {
		if add.Status == addonsv1alpha1.AddonStatusReady {
			addons = append(addons, add)
		}
	}

	return addons
}

func (rc *RepositoryCollection) addonAlreadyRegistered(addon AddonController, addons []addonsv1alpha1.Addon) bool {
	for _, existAddon := range addons {
		if addon.Addon.Name == existAddon.Name && addon.Addon.Version == existAddon.Version {
			return true
		}
	}

	return false
}
