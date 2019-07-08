package repository

import (
	"fmt"

	addonsv1alpha1 "github.com/kyma-project/kyma/components/helm-broker/pkg/apis/addons/v1alpha1"
)

type RepositoryCollection struct {
	Repositories []*RepositoryController
}

func NewRepositoryCollection() *RepositoryCollection {
	return &RepositoryCollection{
		Repositories: []*RepositoryController{},
	}
}

func (rc *RepositoryCollection) AddRepository(repo *RepositoryController) {
	rc.Repositories = append(rc.Repositories, repo)
}

func (rc *RepositoryCollection) Addons() []*AddonController {
	addons := []*AddonController{}

	for _, repo := range rc.Repositories {
		for _, addon := range repo.Addons {
			addons = append(addons, addon)
		}
	}

	return addons
}

func (rc *RepositoryCollection) ReadyAddons() []*AddonController {
	addons := []*AddonController{}

	for _, addon := range rc.Addons() {
		if !addon.IsReady() {
			continue
		}
		addons = append(addons, addon)
	}

	return addons
}

func (rc *RepositoryCollection) IsRepositoriesIdConflict() bool {
	for _, repository := range rc.Repositories {
		if repository.IsFailed() {
			return true
		}
	}

	return false
}

type idConflictData struct {
	repositoryUrl string
	addonsName    string
}

func (rc *RepositoryCollection) ReviseBundleDuplicationInRepository() {
	ids := make(map[string]idConflictData)

	for _, addon := range rc.Addons() {
		if data, ok := ids[addon.ID]; ok {
			addon.ConflictInSpecifiedRepositories(fmt.Errorf("[url: %s, addons: %s]", data.repositoryUrl, data.addonsName))
		} else {
			ids[addon.ID] = idConflictData{
				repositoryUrl: addon.URL,
				addonsName:    fmt.Sprintf("%s:%s", addon.Addon.Name, addon.Addon.Version),
			}
		}
	}
}

func (rc *RepositoryCollection) ReviseBundleDuplicationInStorage(acList *addonsv1alpha1.AddonsConfigurationList) {
	for _, addon := range rc.Addons() {
		rc.findExistingAddon(addon, acList)
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
