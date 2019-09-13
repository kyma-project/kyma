package servicecatalogaddons

import (
	"fmt"
	"strings"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalogaddons/extractor"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/pager"
	"github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"
)

type usageKindService struct {
	informer      cache.SharedIndexInformer
	dynamicClient dynamic.Interface

	extractor extractor.UsageKindUnstructuredExtractor
}

func newUsageKindService(resourceInterface dynamic.Interface, informer cache.SharedIndexInformer) *usageKindService {
	return &usageKindService{
		informer:      informer,
		dynamicClient: resourceInterface,
		extractor:     extractor.UsageKindUnstructuredExtractor{},
	}
}

func (svc *usageKindService) List(params pager.PagingParams) ([]*v1alpha1.UsageKind, error) {
	targets, err := pager.From(svc.informer.GetStore()).Limit(params)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing UsageKinds with paging params [first: %v] [offset: %v]", params.First, params.Offset)
	}

	res := make([]*v1alpha1.UsageKind, 0, len(targets))
	for _, item := range targets {
		u, ok := item.(*unstructured.Unstructured)
		if !ok {
			return nil, fmt.Errorf("incorrect item type: %T, should be: 'UsageKind' in version 'v1alpha1'", item)
		}

		uk, err := svc.extractor.FromUnstructured(u)
		if err != nil {
			return nil, err
		}

		res = append(res, uk)
	}

	return res, nil
}

func (svc *usageKindService) ListResources(namespace string) ([]gqlschema.BindableResourcesOutputItem, error) {
	results := make([]gqlschema.BindableResourcesOutputItem, 0)
	usageKinds := svc.informer.GetStore().List()
	for _, item := range usageKinds {
		uk, err := svc.extractor.Do(item)
		if err != nil {
			return nil, errors.Wrap(err, "while extracting UsageKind")
		}

		ukResources, err := svc.listResourcesForUsageKind(uk, namespace)
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

func (svc *usageKindService) listResourcesForUsageKind(uk *v1alpha1.UsageKind, namespace string) ([]gqlschema.UsageKindResource, error) {
	list, err := svc.dynamicClient.Resource(schema.GroupVersionResource{
		Version:  uk.Spec.Resource.Version,
		Group:    uk.Spec.Resource.Group,
		Resource: strings.ToLower(uk.Spec.Resource.Kind) + "s",
	}).Namespace(namespace).List(v1.ListOptions{})
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
			Namespace: namespace,
		})
	}

	return results, nil
}
