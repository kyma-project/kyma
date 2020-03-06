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
	items := svc.informer.GetStore().List()

	usageKinds, err := svc.extractUsageKinds(items)
	if err != nil {
		return nil, errors.Wrap(err, "while extracting UsageKinds")
	}
	serializedUks := svc.serializeUsageKinds(usageKinds)

	for _, uk := range usageKinds {
		ukResources, err := svc.listResourcesForUsageKind(uk, serializedUks, namespace)
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

func (svc *usageKindService) extractUsageKinds(items []interface{}) ([]*v1alpha1.UsageKind, error) {
	uks := make([]*v1alpha1.UsageKind, 0)
	for _, item := range items {
		uk, err := svc.extractor.Do(item)
		if err != nil {
			return nil, errors.Wrap(err, "while extracting UsageKind")
		}
		uks = append(uks, uk)
	}
	return uks, nil
}

type serializedUsageKind = struct {
	ApiVersion string
	Kind       string
}

func (svc *usageKindService) serializeUsageKinds(uks []*v1alpha1.UsageKind) []serializedUsageKind {
	serialized := make([]serializedUsageKind, 0)
	for _, uk := range uks {
		serialized = append(serialized, serializedUsageKind{
			ApiVersion: fmt.Sprintf("%s/%s", strings.ToLower(uk.Spec.Resource.Group), strings.ToLower(uk.Spec.Resource.Version)),
			Kind:       strings.ToLower(uk.Spec.Resource.Kind),
		})
	}
	return serialized
}

func (svc *usageKindService) listResourcesForUsageKind(uk *v1alpha1.UsageKind, serializedUks []serializedUsageKind, namespace string) ([]gqlschema.UsageKindResource, error) {
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
		if svc.omitResourceByOwnerRefs(serializedUks, item) {
			continue
		}
		results = append(results, gqlschema.UsageKindResource{
			Name:      item.GetName(),
			Namespace: namespace,
		})
	}

	return results, nil
}

func (svc *usageKindService) omitResourceByOwnerRefs(uks []serializedUsageKind, item unstructured.Unstructured) bool {
	for _, uk := range uks {
		for _, ref := range item.GetOwnerReferences() {
			apiVersion := strings.ToLower(ref.APIVersion)
			kind := strings.ToLower(ref.Kind)

			if uk.ApiVersion == apiVersion && uk.Kind == kind {
				return true
			}
			if svc.omitResourceByKServingOwnerRefs(apiVersion, kind) {
				return true
			}
		}
	}
	return false
}

// This function is hardcoded by problem with checking appropriate apiVersion and kind of bindable resource from `serving.knative.dev` apiGroup in `ownerReferences` field of checking resource
func (svc *usageKindService) omitResourceByKServingOwnerRefs(apiVersion, kind string) bool {
	return strings.HasPrefix(apiVersion, "serving.knative.dev") && kind == "revision"
}
