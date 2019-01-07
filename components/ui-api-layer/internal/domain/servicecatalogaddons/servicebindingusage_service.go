package servicecatalogaddons

import (
	"fmt"
	"strings"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/shared"

	api "github.com/kyma-project/kyma/components/binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	"github.com/kyma-project/kyma/components/binding-usage-controller/pkg/client/clientset/versioned/typed/servicecatalog/v1alpha1"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/resource"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

type notifier interface {
	AddListener(observer resource.Listener)
	DeleteListener(observer resource.Listener)
}

type serviceBindingUsageService struct {
	client      v1alpha1.ServicecatalogV1alpha1Interface
	informer    cache.SharedIndexInformer
	scRetriever shared.ServiceCatalogRetriever
	notifier    notifier

	nameFunc func() string
}

func newServiceBindingUsageService(client v1alpha1.ServicecatalogV1alpha1Interface, informer cache.SharedIndexInformer, scRetriever shared.ServiceCatalogRetriever, nameFunc func() string) *serviceBindingUsageService {
	svc := &serviceBindingUsageService{
		client:      client,
		informer:    informer,
		scRetriever: scRetriever,
		nameFunc:    nameFunc,
	}

	informer.AddIndexers(cache.Indexers{
		"usedBy": func(obj interface{}) ([]string, error) {
			serviceBindingUsage, err := svc.toServiceBindingUsage(obj)
			if err != nil {
				return nil, errors.New("while indexing by `usedBy`")
			}

			key := fmt.Sprintf("%s/%s/%s", serviceBindingUsage.Namespace, strings.ToLower(serviceBindingUsage.Spec.UsedBy.Kind), serviceBindingUsage.Spec.UsedBy.Name)

			return []string{key}, nil
		},
	})

	notifier := resource.NewNotifier()
	informer.AddEventHandler(notifier)

	svc.notifier = notifier

	return svc
}

func (f *serviceBindingUsageService) Create(env string, sb *api.ServiceBindingUsage) (*api.ServiceBindingUsage, error) {
	if sb.Name == "" {
		sb.Name = f.nameFunc()
	}
	return f.client.ServiceBindingUsages(env).Create(sb)
}

func (f *serviceBindingUsageService) Delete(env string, name string) error {
	return f.client.ServiceBindingUsages(env).Delete(name, &v1.DeleteOptions{})
}

func (f *serviceBindingUsageService) Find(env string, name string) (*api.ServiceBindingUsage, error) {
	key := fmt.Sprintf("%s/%s", env, name)
	item, exists, err := f.informer.GetStore().GetByKey(key)
	if err != nil || !exists {
		return nil, err
	}

	return f.toServiceBindingUsage(item)
}

func (f *serviceBindingUsageService) List(env string) ([]*api.ServiceBindingUsage, error) {
	items, err := f.informer.GetIndexer().ByIndex("namespace", env)
	if err != nil {
		return nil, err
	}

	return f.toServiceBindingUsages(items)
}

func (f *serviceBindingUsageService) ListForServiceInstance(env string, instanceName string) ([]*api.ServiceBindingUsage, error) {
	bindings, err := f.scRetriever.ServiceBinding().ListForServiceInstance(env, instanceName)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting ServiceBindings for instance [env: %s, name: %s]", env, instanceName)
	}

	bindingNames := make(map[string]struct{})
	for _, binding := range bindings {
		bindingNames[binding.Name] = struct{}{}
	}

	usages, err := f.List(env)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting all ServiceBindingUsages from env: %s", env)
	}
	filteredUsages := make([]*api.ServiceBindingUsage, 0)
	for _, usage := range usages {
		if _, ex := bindingNames[usage.Spec.ServiceBindingRef.Name]; ex {
			filteredUsages = append(filteredUsages, usage)
		}
	}
	return filteredUsages, nil
}

func (f *serviceBindingUsageService) ListForDeployment(environment, kind, deploymentName string) ([]*api.ServiceBindingUsage, error) {
	key := fmt.Sprintf("%s/%s/%s", environment, strings.ToLower(kind), deploymentName)
	indexer := f.informer.GetIndexer()
	items, err := indexer.ByIndex("usedBy", key)
	if err != nil {
		return nil, err
	}

	return f.toServiceBindingUsages(items)
}

func (f *serviceBindingUsageService) Subscribe(listener resource.Listener) {
	f.notifier.AddListener(listener)
}

func (f *serviceBindingUsageService) Unsubscribe(listener resource.Listener) {
	f.notifier.DeleteListener(listener)
}

func (f *serviceBindingUsageService) toServiceBindingUsages(items []interface{}) ([]*api.ServiceBindingUsage, error) {
	var usages []*api.ServiceBindingUsage
	for _, item := range items {
		usage, err := f.toServiceBindingUsage(item)
		if err != nil {
			return nil, err
		}

		usages = append(usages, usage)
	}

	return usages, nil
}

func (f *serviceBindingUsageService) toServiceBindingUsage(item interface{}) (*api.ServiceBindingUsage, error) {
	usage, ok := item.(*api.ServiceBindingUsage)
	if !ok {
		return nil, fmt.Errorf("incorrect item type: %T, should be: *ServiceBindingUsage", item)
	}

	return usage, nil
}
