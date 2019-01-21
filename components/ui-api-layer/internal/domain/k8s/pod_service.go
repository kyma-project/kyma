package k8s

import (
	"errors"
	"fmt"

	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/pager"
	"k8s.io/client-go/tools/cache"
)

type podService struct {
	client   v12.CoreV1Interface
	informer cache.SharedIndexInformer
}

func newPodService(informer cache.SharedIndexInformer, client v12.CoreV1Interface) *podService {
	return &podService{
		client:   client,
		informer: informer,
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

func (svc *podService) Update(name, namespace string, update v1.Pod) (*v1.Pod, error) {
	podKind, podAPIVersion := svc.podTypeMeta()
	if update.Kind != podKind || update.APIVersion != podAPIVersion {
		return nil, errors.New("pod's kind or apiVersion should not be changed")
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

// Kubernetes API used by client-go doesn't provide kind and apiVersion so we have to add it here
// See: https://github.com/kubernetes/kubernetes/issues/3030
func (svc *podService) ensureTypeMeta(pod *v1.Pod) {
	podKind, podAPIVersion := svc.podTypeMeta()
	if pod.APIVersion == "" {
		pod.APIVersion = podAPIVersion
	}
	if pod.Kind == "" {
		pod.Kind = podKind
	}
}

func (svc *podService) podTypeMeta() (kind string, apiVersion string) {
	kind = "Pod"
	apiVersion = "v1"
	return
}
