package k8s

import (
	"fmt"

	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/resource"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/pager"
	"k8s.io/client-go/tools/cache"
)

type podService struct {
	client   corev1.CoreV1Interface
	informer cache.SharedIndexInformer
	notifier resource.Notifier
}

func newPodService(informer cache.SharedIndexInformer, client corev1.CoreV1Interface) *podService {
	notifier := resource.NewNotifier()
	informer.AddEventHandler(notifier)
	return &podService{
		client:   client,
		informer: informer,
		notifier: notifier,
	}
}

func (svc *podService) Find(name, namespace string) (*v1.Pod, error) {
	key := fmt.Sprintf("%s/%s", namespace, name)
	item, exists, err := svc.informer.GetStore().GetByKey(key)
	if err != nil || !exists {
		return nil, err
	}

	pod, ok := item.(*v1.Pod)
	if !ok {
		return nil, fmt.Errorf("Incorrect item type: %T, should be: *Pod", item)
	}

	svc.ensureTypeMeta(pod)

	return pod, nil
}

func (svc *podService) List(namespace string, pagingParams pager.PagingParams) ([]*v1.Pod, error) {
	items, err := pager.FromIndexer(svc.informer.GetIndexer(), "namespace", namespace).Limit(pagingParams)
	if err != nil {
		return nil, err
	}

	var pods []*v1.Pod
	for _, item := range items {
		pod, ok := item.(*v1.Pod)
		if !ok {
			return nil, fmt.Errorf("Incorrect item type: %T, should be: *Pod", item)
		}

		svc.ensureTypeMeta(pod)

		pods = append(pods, pod)
	}

	return pods, nil
}

func (svc *podService) Subscribe(listener resource.Listener) {
	svc.notifier.AddListener(listener)
}

func (svc *podService) Unsubscribe(listener resource.Listener) {
	svc.notifier.DeleteListener(listener)
}

func (svc *podService) Update(name, namespace string, update v1.Pod) (*v1.Pod, error) {
	err := svc.checkUpdatePreconditions(name, namespace, update)
	if err != nil {
		return nil, err
	}

	updated, err := svc.client.Pods(namespace).Update(&update)
	if err != nil {
		return nil, err
	}

	svc.ensureTypeMeta(updated)

	return updated, nil
}

func (svc *podService) Delete(name, namespace string) error {
	return svc.client.Pods(namespace).Delete(name, nil)
}

func (svc *podService) checkUpdatePreconditions(name string, namespace string, update v1.Pod) error {
	errorList := field.ErrorList{}
	if name != update.Name {
		errorList = append(errorList, field.Invalid(field.NewPath("metadata.name"), update.Name, fmt.Sprintf("name of updated object does not match the original (%s)", name)))
	}
	if namespace != update.Namespace {
		errorList = append(errorList, field.Invalid(field.NewPath("metadata.namespace"), update.Namespace, fmt.Sprintf("namespace of updated object does not match the original (%s)", namespace)))
	}
	typeMeta := svc.podTypeMeta()
	if update.Kind != typeMeta.Kind {
		errorList = append(errorList, field.Invalid(field.NewPath("kind"), update.Kind, "pod's kind should not be changed"))
	}
	if update.APIVersion != typeMeta.APIVersion {
		errorList = append(errorList, field.Invalid(field.NewPath("apiVersion"), update.APIVersion, "pod's apiVersion should not be changed"))
	}

	if len(errorList) > 0 {
		return errors.NewInvalid(schema.GroupKind{
			Group: "",
			Kind:  "Pod",
		}, name, errorList)
	}

	return nil
}

// Kubernetes API used by client-go doesn't provide kind and apiVersion so we have to add it here
// See: https://github.com/kubernetes/kubernetes/issues/3030
func (svc *podService) ensureTypeMeta(pod *v1.Pod) {
	pod.TypeMeta = svc.podTypeMeta()
}

func (svc *podService) podTypeMeta() metav1.TypeMeta {
	return metav1.TypeMeta{
		Kind:       "Pod",
		APIVersion: "v1",
	}
}
