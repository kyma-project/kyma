package servicecatalogaddons

import (
	"github.com/kyma-project/helm-broker/pkg/apis/addons/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
)

type clusterAddonsConfigurationConverter struct{}

func (c *clusterAddonsConfigurationConverter) ToGQL(item *v1alpha1.ClusterAddonsConfiguration) *gqlschema.AddonsConfiguration {
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

func (c *clusterAddonsConfigurationConverter) ToGQLs(in []*v1alpha1.ClusterAddonsConfiguration) []gqlschema.AddonsConfiguration {
	var result []gqlschema.AddonsConfiguration
	for _, u := range in {
		converted := c.ToGQL(u)
		if converted != nil {
			result = append(result, *converted)
		}
	}
	return result
}
