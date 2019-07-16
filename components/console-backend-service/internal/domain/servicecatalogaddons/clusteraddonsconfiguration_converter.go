package servicecatalogaddons

import (
	"strings"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/helm-broker/pkg/apis/addons/v1alpha1"
	v1 "k8s.io/api/core/v1"
)

type clusterAddonsConfigurationConverter struct{}

func (c *clusterAddonsConfigurationConverter) ToGQL(item *v1alpha1.ClusterAddonsConfiguration) *gqlschema.AddonsConfiguration {
	if item == nil {
		return nil
	}

	var urls []string
	for _, repo := range item.Spec.Repositories {
		urls = append(urls, repo.URL)
	}

	addonsCfg := gqlschema.AddonsConfiguration{
		Name:   item.Name,
		Labels: item.Labels,
		Urls:   urls,
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

// Deprecated, the ClusterAddonsConfiguration should be used instead
func (c *clusterAddonsConfigurationConverter) ConfigMapToGQL(item *v1.ConfigMap) *gqlschema.AddonsConfiguration {
	if item == nil {
		return nil
	}

	var urls []string
	if len(item.Data["URLs"]) > 0 {
		urls = strings.Split(item.Data["URLs"], "\n")
	}

	addonsCfg := gqlschema.AddonsConfiguration{
		Name:   item.Name,
		Labels: item.Labels,
		Urls:   urls,
	}

	return &addonsCfg
}

// Deprecated, the ClusterAddonsConfiguration should be used instead
func (c *clusterAddonsConfigurationConverter) ConfigMapToGQLs(in []*v1.ConfigMap) []gqlschema.AddonsConfiguration {
	var result []gqlschema.AddonsConfiguration
	for _, u := range in {
		converted := c.ConfigMapToGQL(u)
		if converted != nil {
			result = append(result, *converted)
		}
	}
	return result
}
