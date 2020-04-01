package finder

import (
	"fmt"

	newapi "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	oldapi "github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma-project.io/v1alpha2"
	"github.com/kyma-project/kyma/components/api-gateway-migrator/pkg/migrator"
	log "github.com/sirupsen/logrus"
)

type filter func(oldApi *oldapi.Api) (bool, string)

var (
	invalidStatusFilter = func(oldApi *oldapi.Api) (bool, string) {
		//TODO: Implement
		return false, ""
	}

	labelFilter = func(omitApisWithLabels map[string]*string) func(oldApi *oldapi.Api) (bool, string) {
		return func(oldApi *oldapi.Api) (bool, string) {
			for key, value := range omitApisWithLabels {
				if objectLabelValue, exists := oldApi.Labels[key]; exists {
					if value == nil || objectLabelValue == *value {
						return true, fmt.Sprintf("object matches configured label: %s", key)
					}
				}
			}
			return false, ""
		}
	}

	doubleJwtFilter = func(oldApi *oldapi.Api) (bool, string) {
		//TODO: Implement
		return false, ""
	}
)

type Finder struct {
	filters   []filter
	k8sClient *migrator.K8sClient
}

func (f *Finder) withFilters(filters []filter) *Finder {
	f.filters = filters
	return f
}

func (f *Finder) withClient(k8sClient *migrator.K8sClient) *Finder {
	f.k8sClient = k8sClient
	return f
}

func New(k8sClient *migrator.K8sClient, labels map[string]*string) *Finder {

	return (&Finder{}).
		withFilters([]filter{invalidStatusFilter, labelFilter(labels), doubleJwtFilter}).
		withClient(k8sClient)
}

func (f *Finder) Find() ([]oldapi.Api, error) {

	//List old api objects: apis.gateway.kyma-project.io/v1alpha2
	oldApiList, err := f.k8sClient.FindOldApis()
	if err != nil {
		return nil, err
	}

	res := []oldapi.Api{}

	for _, oldApi := range oldApiList {

		shouldSkip := false
		for _, filter := range f.filters {
			ol := oldApi
			ss, reason := filter(&ol)
			if ss {
				log.Infof("Skipping migration for Api: \"%s\" in namespace: \"%s\". Reason: \"%s\"", oldApi.Name, oldApi.Namespace, reason)
				shouldSkip = true
				break
			}
		}

		if !shouldSkip {
			res = append(res, oldApi)
		}
	}

	return res, nil
}

func shouldMigrate(oldApi oldapi.Api, newApis []newapi.APIRule, omitApisWithLabels map[string]*string) bool {
	for key, value := range omitApisWithLabels {
		if objectLabelValue, exists := oldApi.Labels[key]; exists {
			if value == nil || objectLabelValue == *value {
				return false
			}
		}
	}
	return true
}
