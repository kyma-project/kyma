package servicecatalog

import (
	"fmt"

	"github.com/kyma-project/kyma/components/binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	v1alpha12 "github.com/kyma-project/kyma/components/binding-usage-controller/pkg/client/clientset/versioned/typed/servicecatalog/v1alpha1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/pager"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"
)

type usageKindDynamicOperations interface {
	ClientForGroupVersionKind(kind schema.GroupVersionKind) (dynamic.Interface, error)
}

type usageKindService struct {
	informer      cache.SharedIndexInformer
	dynamicClient usageKindDynamicOperations
	client        v1alpha12.ServicecatalogV1alpha1Interface
}

func newUsageKindService(client v1alpha12.ServicecatalogV1alpha1Interface, dynamicClient usageKindDynamicOperations, informer cache.SharedIndexInformer) *usageKindService {
	return &usageKindService{
		informer:      informer,
		dynamicClient: dynamicClient,
		client:        client,
	}
}

func (svc *usageKindService) List(params pager.PagingParams) ([]*v1alpha1.UsageKind, error) {
	targets, err := pager.From(svc.informer.GetStore()).Limit(params)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing UsageKinds with paging params [first: %v] [offset: %v]: %v", params.First, params.Offset)
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

func (svc *usageKindService) ListUsageKindResources(usageKind string, environment string) ([]gqlschema.UsageKindResource, error) {
	target, err := svc.client.UsageKinds().Get(usageKind, v1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "while getting UsageKind: %s", usageKind)
	}

	cli, err := svc.dynamicClient.ClientForGroupVersionKind(schema.GroupVersionKind{
		Kind:    target.Spec.Resource.Kind,
		Group:   target.Spec.Resource.Group,
		Version: target.Spec.Resource.Version,
	})
	if err != nil {
		return nil, errors.Wrap(err, "while creating dynamic client")
	}

	client := cli.Resource(&v1.APIResource{
		Namespaced: true,
		Name:       target.Spec.Resource.Kind + "s",
	}, environment)

	obj, err := client.List(v1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "while listing target resources")
	}

	list, ok := obj.(*unstructured.UnstructuredList)
	if !ok {
		return nil, errors.New("cannot list object to UnstructuredList")
	}

	results := make([]gqlschema.UsageKindResource, 0)
	for _, item := range list.Items {
		// TODO: Write test case for logic bellow
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
