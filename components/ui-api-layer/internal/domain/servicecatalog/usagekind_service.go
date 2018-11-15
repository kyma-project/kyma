package servicecatalog

import (
	"fmt"
	"strings"

	"github.com/kyma-project/kyma/components/binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	v1alpha12 "github.com/kyma-project/kyma/components/binding-usage-controller/pkg/client/clientset/versioned/typed/servicecatalog/v1alpha1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/pager"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"
)

type usageKindService struct {
	informer      cache.SharedIndexInformer
	dynamicClient dynamic.Interface
	scClient      v1alpha12.ServicecatalogV1alpha1Interface
}

func newUsageKindService(client v1alpha12.ServicecatalogV1alpha1Interface, resourceInterface dynamic.Interface, informer cache.SharedIndexInformer) *usageKindService {
	return &usageKindService{
		informer:      informer,
		dynamicClient: resourceInterface,
		scClient:      client,
	}
}

func (svc *usageKindService) List(params pager.PagingParams) ([]*v1alpha1.UsageKind, error) {
	targets, err := pager.From(svc.informer.GetStore()).Limit(params)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing UsageKinds with paging params [first: %v] [offset: %v]", params.First, params.Offset)
	}

	res := make([]*v1alpha1.UsageKind, 0, len(targets))
	for _, item := range targets {
		uk, ok := item.(*v1alpha1.UsageKind)
		if !ok {
			return nil, fmt.Errorf("incorrect item type: %T, should be: 'UsageKind' in version 'v1alpha1'", item)
		}

		res = append(res, uk)
	}

	return res, nil
}

func (svc *usageKindService) ListResources(environment string) ([]gqlschema.BindableResourcesOutputItem, error) {
	results := make([]gqlschema.BindableResourcesOutputItem, 0)
	usageKinds := svc.informer.GetStore().List()
	for _, item := range usageKinds {
		uk, ok := item.(*v1alpha1.UsageKind)
		if !ok {
			return nil, fmt.Errorf("incorrect item type: %T, should be: *UsageKind", item)
		}

		ukResources, err := svc.listResourcesForUsageKind(uk, environment)
		if err != nil {
			return nil, errors.Wrap(err, "while listing target resources")
		}

		results = append(results, gqlschema.BindableResourcesOutputItem{
			Kind:        uk.Name,
			DisplayName: uk.Spec.DisplayName,
			Resources:   ukResources,
		})
	}
	return results, nil
}

func (svc *usageKindService) ListUsageKindResources(usageKind string, environment string) ([]gqlschema.UsageKindResource, error) {
	target, err := svc.scClient.UsageKinds().Get(usageKind, v1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "while getting UsageKind: %s", usageKind)
	}

	return svc.listResourcesForUsageKind(target, environment)
}

func (svc *usageKindService) listResourcesForUsageKind(uk *v1alpha1.UsageKind, environment string) ([]gqlschema.UsageKindResource, error) {
	list, err := svc.dynamicClient.Resource(schema.GroupVersionResource{
		Version:  uk.Spec.Resource.Version,
		Group:    uk.Spec.Resource.Group,
		Resource: strings.ToLower(uk.Spec.Resource.Kind) + "s",
	}).Namespace(environment).List(v1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "while listing target resources")
	}

	results := make([]gqlschema.UsageKindResource, 0)
	for _, item := range list.Items {
		if len(item.GetOwnerReferences()) > 0 {
			continue
		}
		results = append(results, gqlschema.UsageKindResource{
			Name:      item.GetName(),
			Namespace: environment,
		})
	}

	return results, nil
}
