package finder

import (
	"fmt"

	oldapiMeta "github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma-project.io/meta/v1"
	oldapi "github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma-project.io/v1alpha2"
	"github.com/kyma-project/kyma/components/api-gateway-migrator/pkg/migrator"
	log "github.com/sirupsen/logrus"
)

type filter func(oldApi *oldapi.Api) (bool, string)

var (
	invalidStatusFilter = newStatusFilter()
	labelFilter         = newLabelFilter
	doubleJwtFilter     = newJwtStructureFilter()
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

	res := f.filterApis(oldApiList)

	return res, nil
}

func (f *Finder) filterApis(oldApiList []oldapi.Api) []oldapi.Api {

	res := []oldapi.Api{}

	for _, oldApi := range oldApiList {

		shouldSkip := false
		for _, filter := range f.filters {
			ol := oldApi
			filtRes, reason := filter(&ol)
			if filtRes {
				log.Infof("Skipping migration for Api: \"%s\" in namespace: \"%s\". Reason: \"%s\"", oldApi.Name, oldApi.Namespace, reason)
				shouldSkip = true
				break
			}
		}

		if !shouldSkip {
			res = append(res, oldApi)
		}
	}

	return res
}

func newLabelFilter(omitApisWithLabels map[string]*string) func(oldApi *oldapi.Api) (bool, string) {

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

func newStatusFilter() func(oldApi *oldapi.Api) (bool, string) {

	return func(oldApi *oldapi.Api) (bool, string) {
		if oldApi.Status.ValidationStatus != oldapiMeta.Successful {
			return true, fmt.Sprintf("Invalid validationStatus code: %d", oldApi.Status.ValidationStatus)
		}

		if oldApi.Status.VirtualServiceStatus.Code != oldapiMeta.Successful {
			return true, fmt.Sprintf("Invalid virtualServiceStatus code: %d", oldApi.Status.VirtualServiceStatus.Code)
		}

		if oldApi.Status.AuthenticationStatus.Code != oldapiMeta.Successful {
			return true, fmt.Sprintf("Invalid authenticationStatus code: %d", oldApi.Status.AuthenticationStatus.Code)
		}

		return false, ""
	}
}

func newJwtStructureFilter() func(oldApi *oldapi.Api) (bool, string) {
	return func(oldApi *oldapi.Api) (bool, string) {
		if len(oldApi.Spec.Authentication) < 2 {
			return false, ""
		}

		allowedTriggerRules := oldApi.Spec.Authentication[0].Jwt.TriggerRule

		for i := 1; i < len(oldApi.Spec.Authentication); i++ {
			if areDifferent(allowedTriggerRules, oldApi.Spec.Authentication[i].Jwt.TriggerRule) {
				return true, "object is configured with more than one jwt authentication that contain different triggerRules"
			}
		}

		return false, ""
	}
}

func areDifferent(first, second *oldapi.TriggerRule) bool {
	firstPathsCount := 0
	if first != nil {
		firstPathsCount = len(first.ExcludedPaths)
	}

	secondPathsCount := 0
	if second != nil {
		secondPathsCount = len(second.ExcludedPaths)
	}

	if firstPathsCount != secondPathsCount {
		return true
	}

	//At this point we know counts are the same. Is there a need to iterate?
	if firstPathsCount > 0 {

		//Brute-force approach. We don't require same order, just the same set of items.

		//Iterate over items from first object
		for fi := 0; fi < firstPathsCount; fi++ {

			fItem := first.ExcludedPaths[fi]
			matchFound := false

			//Try to find a match in the items from second object
			for si := 0; si < secondPathsCount; si++ {
				sItem := second.ExcludedPaths[si]
				if fItem.ExprType == sItem.ExprType && fItem.Value == sItem.Value {
					matchFound = true
					break
				}
			}

			if !matchFound {
				return true
			}
		}
	}

	return false
}
