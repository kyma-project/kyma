package k8s

import (
	"fmt"

	"github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/pager"
	"k8s.io/client-go/tools/cache"
)

type configMapService struct {
	client   corev1.CoreV1Interface
	informer cache.SharedIndexInformer
}

func newConfigMapService(informer cache.SharedIndexInformer, client corev1.CoreV1Interface) *configMapService {
	notifier := resource.NewNotifier()
	informer.AddEventHandler(notifier)
	return &configMapService{
		client:   client,
		informer: informer,
	}
}

func (svc *configMapService) Find(name, namespace string) (*v1.ConfigMap, error) {
	key := fmt.Sprintf("%s/%s", namespace, name)
	item, exists, err := svc.informer.GetStore().GetByKey(key)
	if err != nil || !exists {
		return nil, err
	}

	configMap, ok := item.(*v1.ConfigMap)
	if !ok {
		return nil, fmt.Errorf("Incorrect item type: %T, should be: *ConfigMap", item)
	}

	svc.ensureTypeMeta(configMap)

	return configMap, nil
}

func (svc *configMapService) List(namespace string, pagingParams pager.PagingParams) ([]*v1.ConfigMap, error) {
	items, err := pager.FromIndexer(svc.informer.GetIndexer(), "namespace", namespace).Limit(pagingParams)
	if err != nil {
		return nil, err
	}

	var configMaps []*v1.ConfigMap
	for _, item := range items {
		configMap, ok := item.(*v1.ConfigMap)
		if !ok {
			return nil, fmt.Errorf("Incorrect item type: %T, should be: *ConfigMap", item)
		}

		svc.ensureTypeMeta(configMap)

		configMaps = append(configMaps, configMap)
	}

	return configMaps, nil
}

func (svc *configMapService) Update(name, namespace string, update v1.ConfigMap) (*v1.ConfigMap, error) {
	err := svc.checkUpdatePreconditions(name, namespace, update)
	if err != nil {
		return nil, err
	}

	updated, err := svc.client.ConfigMaps(namespace).Update(&update)
	if err != nil {
		return nil, err
	}

	svc.ensureTypeMeta(updated)

	return updated, nil
}

func (svc *configMapService) Delete(name, namespace string) error {
	return svc.client.ConfigMaps(namespace).Delete(name, nil)
}

func (svc *configMapService) checkUpdatePreconditions(name string, namespace string, update v1.ConfigMap) error {
	errorList := field.ErrorList{}
	if name != update.Name {
		errorList = append(errorList, field.Invalid(field.NewPath("metadata.name"), update.Name, fmt.Sprintf("name of updated object does not match the original (%s)", name)))
	}
	if namespace != update.Namespace {
		errorList = append(errorList, field.Invalid(field.NewPath("metadata.namespace"), update.Namespace, fmt.Sprintf("namespace of updated object does not match the original (%s)", namespace)))
	}
	typeMeta := svc.configMapTypeMeta()
	if update.Kind != typeMeta.Kind {
		errorList = append(errorList, field.Invalid(field.NewPath("kind"), update.Kind, "ConfigMap's kind should not be changed"))
	}
	if update.APIVersion != typeMeta.APIVersion {
		errorList = append(errorList, field.Invalid(field.NewPath("apiVersion"), update.APIVersion, "ConfigMap's apiVersion should not be changed"))
	}

	if len(errorList) > 0 {
		return errors.NewInvalid(schema.GroupKind{
			Group: "",
			Kind:  "ConfigMap",
		}, name, errorList)
	}

	return nil
}

// Kubernetes API used by client-go doesn't provide kind and apiVersion so we have to add it here
// See: https://github.com/kubernetes/kubernetes/issues/3030
func (svc *configMapService) ensureTypeMeta(configMap *v1.ConfigMap) {
	configMap.TypeMeta = svc.configMapTypeMeta()
}

func (svc *configMapService) configMapTypeMeta() metav1.TypeMeta {
	return metav1.TypeMeta{
		Kind:       "ConfigMap",
		APIVersion: "apps/v1",
	}
}
