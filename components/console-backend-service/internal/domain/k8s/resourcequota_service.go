package k8s

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	coreV1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
)

type resourceQuotaService struct {
	rqInformer cache.SharedIndexInformer
	rsInformer cache.SharedIndexInformer
	ssInformer cache.SharedIndexInformer
	podClient  coreV1.CoreV1Interface
}

func newResourceQuotaService(rqInformer cache.SharedIndexInformer, rsInformer cache.SharedIndexInformer, ssInformer cache.SharedIndexInformer, podClient coreV1.CoreV1Interface) *resourceQuotaService {
	return &resourceQuotaService{
		rqInformer: rqInformer,
		rsInformer: rsInformer,
		ssInformer: ssInformer,
		podClient:  podClient,
	}
}

func (svc *resourceQuotaService) ListResourceQuotas(namespace string) ([]*v1.ResourceQuota, error) {
	items, err := svc.rqInformer.GetIndexer().ByIndex(cache.NamespaceIndex, namespace)
	if err != nil {
		return []*v1.ResourceQuota{}, err
	}

	var result []*v1.ResourceQuota
	for _, item := range items {
		rq, ok := item.(*v1.ResourceQuota)
		if !ok {
			return nil, fmt.Errorf("unexpected item type: %T, should be *ResourceQuota", item)
		}
		result = append(result, rq)
	}

	return result, nil
}

func (svc *resourceQuotaService) ListReplicaSets(namespace string) ([]*apps.ReplicaSet, error) {
	items, err := svc.rsInformer.GetIndexer().ByIndex(cache.NamespaceIndex, namespace)
	if err != nil {
		return []*apps.ReplicaSet{}, err
	}

	var result []*apps.ReplicaSet
	for _, item := range items {
		rq, ok := item.(*apps.ReplicaSet)
		if !ok {
			return nil, fmt.Errorf("unexpected item type: %T, should be *ResourceQuota", item)
		}
		result = append(result, rq)
	}

	return result, nil
}

func (svc *resourceQuotaService) ListStatefulSets(namespace string) ([]*apps.StatefulSet, error) {
	items, err := svc.ssInformer.GetIndexer().ByIndex(cache.NamespaceIndex, namespace)
	if err != nil {
		return []*apps.StatefulSet{}, err
	}

	var result []*apps.StatefulSet
	for _, item := range items {
		rq, ok := item.(*apps.StatefulSet)
		if !ok {
			return nil, fmt.Errorf("unexpected item type: %T, should be *ResourceQuota", item)
		}
		result = append(result, rq)
	}

	return result, nil
}

func (svc *resourceQuotaService) ListPods(namespace string, labelSelector map[string]string) ([]v1.Pod, error) {
	selectors := make([]string, 0)
	for key, value := range labelSelector {
		selectors = append(selectors, fmt.Sprintf("%s=%s", key, value))
	}

	selector := strings.Join(selectors, ",")

	pods, err := svc.podClient.Pods(namespace).List(metaV1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "while listing Pods in namespace: %selector, with labelSelector: %selector", namespace, labelSelector)
	}

	return pods.Items, err
}
