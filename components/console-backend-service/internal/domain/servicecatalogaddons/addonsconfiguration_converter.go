package servicecatalogaddons

import (
	"github.com/kyma-project/helm-broker/pkg/apis/addons/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
)

type addonsConfigurationConverter struct{}

func (c *addonsConfigurationConverter) ToGQL(item *v1alpha1.AddonsConfiguration) *gqlschema.AddonsConfiguration {
	if item == nil {
		return nil
	}

	var urls []string
	var repositories []gqlschema.AddonsConfigurationRepository
	for _, repo := range item.Spec.Repositories {
		urls = append(urls, repo.URL)
		repositories = append(repositories, parseRepository(repo))
	}

	addonsCfg := gqlschema.AddonsConfiguration{
		Name:         item.Name,
		Labels:       item.Labels,
		Urls:         urls,
		Status:       parseStatus(item.Status.CommonAddonsConfigurationStatus),
		Repositories: repositories,
	}

	return &addonsCfg
}

func (c *addonsConfigurationConverter) ToGQLs(in []*v1alpha1.AddonsConfiguration) []gqlschema.AddonsConfiguration {
	var result []gqlschema.AddonsConfiguration
	for _, u := range in {
		converted := c.ToGQL(u)
		if converted != nil {
			result = append(result, *converted)
		}
	}
	return result
}

func parseRepository(repo v1alpha1.SpecRepository) gqlschema.AddonsConfigurationRepository {
	secretRef := &gqlschema.ResourceRef{}
	if repo.SecretRef != nil {
		secretRef.Name = repo.SecretRef.Name
		secretRef.Namespace = repo.SecretRef.Namespace
	} else {
		secretRef = nil
	}
	return gqlschema.AddonsConfigurationRepository{
		URL:       repo.URL,
		SecretRef: secretRef,
	}
}

func parseStatus(status v1alpha1.CommonAddonsConfigurationStatus) gqlschema.AddonsConfigurationStatus {
	var repositories []gqlschema.AddonsConfigurationStatusRepository
	for _, repo := range status.Repositories {
		var addons []gqlschema.AddonsConfigurationStatusAddons
		for _, addon := range repo.Addons {
			addons = append(addons, gqlschema.AddonsConfigurationStatusAddons{
				Status:  string(addon.Status),
				Name:    addon.Name,
				Version: addon.Version,
				Message: addon.Message,
				Reason:  string(addon.Reason),
			})
		}
		repositories = append(repositories, gqlschema.AddonsConfigurationStatusRepository{
			Status:  string(repo.Status),
			URL:     repo.URL,
			Addons:  addons,
			Message: repo.Message,
			Reason:  string(repo.Reason),
		})
	}
	return gqlschema.AddonsConfigurationStatus{
		Phase:        string(status.Phase),
		Repositories: repositories,
	}
}
