package apicontroller

import (
	"fmt"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/apicontroller/extractor"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"

	"github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma-project.io/v1alpha2"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"
	"github.com/pkg/errors"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"
)

type apiService struct {
	informer  cache.SharedIndexInformer
	client    dynamic.NamespaceableResourceInterface
	notifier  resource.Notifier
	extractor extractor.ApiUnstructuredExtractor
}

var apisTypeMeta = metav1.TypeMeta{
	Kind:       "Api",
	APIVersion: "gateway.kyma-project.io/v1alpha2",
}

func newApiService(informer cache.SharedIndexInformer, client dynamic.NamespaceableResourceInterface) *apiService {
	notifier := resource.NewNotifier()
	informer.AddEventHandler(notifier)

	return &apiService{
		informer:  informer,
		client:    client,
		notifier:  notifier,
		extractor: extractor.ApiUnstructuredExtractor{},
	}
}

func (svc *apiService) List(namespace string, serviceName *string, hostname *string) ([]*v1alpha2.Api, error) {
	items, err := svc.informer.GetIndexer().ByIndex("namespace", namespace)
	if err != nil {
		return nil, err
	}

	var apis []*v1alpha2.Api
	for _, item := range items {

		api, ok := item.(*unstructured.Unstructured)
		if !ok {
			return nil, fmt.Errorf("incorrect item type: %T, should be: *unstructured.Unstructured", item)
		}

		formattedApi, err := svc.extractor.FromUnstructured(api)
		if err != nil {
			return nil, err
		}
		match := true
		if serviceName != nil {
			if *serviceName != formattedApi.Spec.Service.Name {
				match = false
			}
		}
		if hostname != nil {
			if *hostname != formattedApi.Spec.Hostname {
				match = false
			}
		}

		if match {
			apis = append(apis, formattedApi)
		}
	}

	return apis, nil
}

func (svc *apiService) Find(name string, namespace string) (*v1alpha2.Api, error) {
	key := fmt.Sprintf("%s/%s", namespace, name)
	item, exists, err := svc.informer.GetStore().GetByKey(key)
	if err != nil {
		return nil, errors.Wrapf(err, "while finding API %s", name)
	}

	if !exists {
		return nil, nil
	}

	res, err := svc.extractor.Do(item)
	if err != nil {
		return &v1alpha2.Api{}, err
	}

	return res, nil
}

func (svc *apiService) Create(api *v1alpha2.Api) (*v1alpha2.Api, error) {
	api.TypeMeta = apisTypeMeta

	u, err := svc.extractor.ToUnstructured(api)
	if err != nil {
		return &v1alpha2.Api{}, err
	}

	created, err := svc.client.Namespace(api.ObjectMeta.Namespace).Create(u, metav1.CreateOptions{})
	if err != nil {
		return &v1alpha2.Api{}, err
	}
	return svc.extractor.FromUnstructured(created)
}

func (svc *apiService) Subscribe(listener resource.Listener) {
	svc.notifier.AddListener(listener)
}

func (svc *apiService) Unsubscribe(listener resource.Listener) {
	svc.notifier.DeleteListener(listener)
}

func (svc *apiService) Update(api *v1alpha2.Api) (*v1alpha2.Api, error) {
	oldApi, err := svc.Find(api.ObjectMeta.Name, api.ObjectMeta.Namespace)
	if err != nil {
		return nil, errors.Wrapf(err, "while finding API %s", api.ObjectMeta.Name)
	}

	if oldApi == nil {
		return nil, apiErrors.NewNotFound(schema.GroupResource{
			Group:    apisGroupVersionResource.Group,
			Resource: apisGroupVersionResource.Resource,
		}, api.ObjectMeta.Name)
	}
	api.ObjectMeta.ResourceVersion = oldApi.ObjectMeta.ResourceVersion
	api.TypeMeta = apisTypeMeta

	u, err := svc.extractor.ToUnstructured(api)
	if err != nil {
		return &v1alpha2.Api{}, err
	}

	updated, err := svc.client.Namespace(api.ObjectMeta.Namespace).Update(u, metav1.UpdateOptions{})
	if err != nil {
		return &v1alpha2.Api{}, err
	}
	return svc.extractor.FromUnstructured(updated)
}

func (svc *apiService) Delete(name string, namespace string) error {
	return svc.client.Namespace(namespace).Delete(name, nil)
}
