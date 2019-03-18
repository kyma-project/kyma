package k8s

import (
	"fmt"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/pager"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
)

type serviceService struct {
	client   corev1.CoreV1Interface
	informer cache.SharedIndexInformer
	notifier resource.Notifier
}

func newServiceService(informer cache.SharedIndexInformer) *serviceService {
	notifier := resource.NewNotifier()
	informer.AddEventHandler(notifier)
	return &serviceService{
		informer: informer,
		notifier: notifier,
	}
}

func (svc *serviceService) List(namespace string, pagingParams pager.PagingParams) ([]*v1.Service, error) {
	items, err := pager.FromIndexer(svc.informer.GetIndexer(), "namespace", namespace).Limit(pagingParams)
	if err != nil {
		return nil, err
	}
	services := make([]*v1.Service, len(items))
	for i, item := range items {
		service, ok := item.(*v1.Service)
		if !ok {
			return nil, fmt.Errorf("incorrect item type: %T, should be: *Service", item)
		}
		service.TypeMeta = metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		}
		services[i] = service
	}

	return services, nil
}

func (svc *serviceService) Find(name, namespace string) (*v1.Service, error) {
	key := fmt.Sprintf("%s/%s", namespace, name)

	item, exists, err := svc.informer.GetStore().GetByKey(key)
	if err != nil || !exists {
		return nil, err
	}

	service, ok := item.(*v1.Service)
	if !ok {
		return nil, fmt.Errorf("incorrect item type: %T, should be: *v1.Service", item)
	}
	svc.ensureTypeMeta(service)
	return service, nil
}

func (svc *serviceService) ensureTypeMeta(service *v1.Service) {
	service.TypeMeta = svc.serviceTypeMetadata()
}

func (svc *serviceService) serviceTypeMetadata() metav1.TypeMeta {
	return metav1.TypeMeta{
		Kind:       "Service",
		APIVersion: "v1",
	}
}

func (svc *serviceService) Subscribe(listener resource.Listener) {
	svc.notifier.AddListener(listener)
}

func (svc *serviceService) Unsubscribe(listener resource.Listener) {
	svc.notifier.DeleteListener(listener)
}

func (svc *serviceService) Update(name, namespace string, update v1.Service) (*v1.Service, error) {
	err := svc.checkUpdatePreconditions(name, namespace, update)
	if err != nil {
		return nil, err
	}

	updated, err := svc.client.Services(namespace).Update(&update)
	if err != nil {
		return nil, err
	}

	svc.ensureTypeMeta(updated)

	return updated, nil
}

func (svc *serviceService) Delete(name, namespace string) error {
	return svc.client.Services(namespace).Delete(name, nil)
}

func (svc *serviceService) checkUpdatePreconditions(name string, namespace string, update v1.Service) error {
	errorList := field.ErrorList{}
	if name != update.Name {
		errorList = append(errorList, field.Invalid(field.NewPath("metadata.name"), update.Name, fmt.Sprintf("name of updated object does not match the original (%s)", name)))
	}
	if namespace != update.Namespace {
		errorList = append(errorList, field.Invalid(field.NewPath("metadata.namespace"), update.Namespace, fmt.Sprintf("namespace of updated object does not match the original (%s)", namespace)))
	}
	typeMeta := svc.serviceTypeMetadata()
	if update.Kind != typeMeta.Kind {
		errorList = append(errorList, field.Invalid(field.NewPath("kind"), update.Kind, "service kind should not be changed"))
	}
	if update.APIVersion != typeMeta.APIVersion {
		errorList = append(errorList, field.Invalid(field.NewPath("apiVersion"), update.APIVersion, "service apiVersion should not be changed"))
	}

	if len(errorList) > 0 {
		return errors.NewInvalid(schema.GroupKind{
			Group: "",
			Kind:  "Service",
		}, name, errorList)
	}

	return nil
}
