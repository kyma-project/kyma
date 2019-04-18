package servicecatalogaddons

import (
	"strings"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	v1 "k8s.io/api/core/v1"
)

type addonsConfigurationConverter struct{}

func (c *addonsConfigurationConverter) ToGQL(item *v1.ConfigMap) *gqlschema.AddonsConfiguration {
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

func (c *addonsConfigurationConverter) ToGQLs(in []*v1.ConfigMap) []gqlschema.AddonsConfiguration {
	var result []gqlschema.AddonsConfiguration
	for _, u := range in {
		converted := c.ToGQL(u)
		if converted != nil {
			result = append(result, *converted)
		}
	}
	return result
}
