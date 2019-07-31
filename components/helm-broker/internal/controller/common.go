package controller

import (
	"fmt"

	add "github.com/kyma-project/kyma/components/helm-broker/internal/addon"
	"github.com/kyma-project/kyma/components/helm-broker/internal/controller/addons"
	"github.com/kyma-project/kyma/components/helm-broker/pkg/apis/addons/v1alpha1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type addonLoader struct {
	addonGetterFactory addonGetterFactory
	log                logrus.FieldLogger
	dstPath            string
}

// Load loads repositories from given addon
func (a *addonLoader) Load(repos []v1alpha1.SpecRepository) *addons.RepositoryCollection {
	repositories := addons.NewRepositoryCollection()
	for _, specRepository := range repos {
		a.log.Infof("- create addons for %q repository", specRepository.URL)
		repo := addons.NewAddonsRepository(specRepository.URL)

		adds, err := a.createAddons(specRepository.URL)
		if err != nil {
			repo.FetchingError(err)
			repositories.AddRepository(repo)

			a.log.Errorf("while creating addons for repository from %q: %s", specRepository.URL, err)
			continue
		}

		repo.Addons = adds
		repositories.AddRepository(repo)
	}
	return repositories
}

func (a *addonLoader) createAddons(URL string) ([]*addons.AddonController, error) {
	concreteGetter, err := a.addonGetterFactory.NewGetter(URL, a.dstPath)
	if err != nil {
		return nil, err
	}
	defer concreteGetter.Cleanup()

	// fetch repository index
	index, err := concreteGetter.GetIndex()
	if err != nil {
		return nil, errors.Wrap(err, "while reading repository index")
	}

	// for each repository entry create addon
	var adds []*addons.AddonController
	for _, entries := range index.Entries {
		for _, entry := range entries {
			addon := addons.NewAddon(string(entry.Name), string(entry.Version), URL)
			adds = append(adds, addon)

			completeAddon, err := concreteGetter.GetCompleteAddon(entry)
			switch {
			case err == nil:
				addon.ID = string(completeAddon.Addon.ID)
				addon.CompleteAddon = completeAddon.Addon
				addon.Charts = completeAddon.Charts
			case add.IsFetchingError(err):
				addon.FetchingError(err)
				a.log.WithField("addon", fmt.Sprintf("%s-%s", entry.Name, entry.Version)).Errorf("while fetching addon: %s", err)
			default:
				addon.LoadingError(err)
				a.log.WithField("addon", fmt.Sprintf("%s-%s", entry.Name, entry.Version)).Errorf("while loading addon: %s", err)
			}
		}
	}

	return adds, nil
}
