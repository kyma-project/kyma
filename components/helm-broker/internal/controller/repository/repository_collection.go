package repository

import (
	"fmt"

	addonsv1alpha1 "github.com/kyma-project/kyma/components/helm-broker/pkg/apis/addons/v1alpha1"
)

type RepositoryCollection struct {
	repositories []*RepositoryController
}

func NewRepositoryCollection() *RepositoryCollection {
	return &RepositoryCollection{repositories: []*RepositoryController{}}
}

func (rc *RepositoryCollection) AddRepository(repo *RepositoryController) {
	rc.repositories = append(rc.repositories, repo)
}

func (rc *RepositoryCollection) Collection() []*RepositoryController {
	return rc.repositories
}

type idConflictData struct {
	repositoryUrl string
	addonsName    string
}

func (rc *RepositoryCollection) IsReady() bool {
	for _, repo := range rc.Collection() {
		if !repo.IsReady() {
			return false
		}
	}

	return true
}

func (rc *RepositoryCollection) ReviseBundleDuplicationInRepository() {
	ids := make(map[string]idConflictData)

	for _, repo := range rc.Collection() {
		for _, addon := range repo.Addons {
			if data, ok := ids[addon.ID]; ok {
				addon.Failed()
				addon.SetAddonFailedInfo(
					addonsv1alpha1.AddonConflictInSpecifiedRepositories,
					fmt.Sprintf("[url: %s, addons: %s]", data.repositoryUrl, data.addonsName),
				)
			} else {
				ids[addon.ID] = idConflictData{
					repositoryUrl: repo.Repository.URL,
					addonsName:    fmt.Sprintf("%s:%s", addon.Addon.Name, addon.Addon.Version),
				}
			}
		}
	}
}

func (rc *RepositoryCollection) ReviseBundleDuplicationInStorage(acList *addonsv1alpha1.AddonsConfigurationList) {
	for _, repo := range rc.repositories {
		for _, addon := range repo.Addons {
			rc.findExistingAddon(addon, acList)
		}
	}
}

func (rc *RepositoryCollection) findExistingAddon(addon *AddonController, list *addonsv1alpha1.AddonsConfigurationList) {
	for _, existAc := range list.Items {
		for _, repo := range existAc.Status.Repositories {
			if rc.addonAlreadyRegistered(*addon, rc.filterReadyAddons(repo)) {
				addon.Failed()
				addon.SetAddonFailedInfo(
					addonsv1alpha1.AddonConflictWithAlreadyRegisteredAddons,
					fmt.Sprintf("[ConfigurationName: %s, url: %s, addons: %s:%s]", existAc.Name, repo.URL, addon.Addon.Name, addon.Addon.Version),
				)
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
