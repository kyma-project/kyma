package k8s

import (
	"fmt"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/pager"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"
	v1 "k8s.io/api/core/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

type secretService struct {
	client   corev1.CoreV1Interface
	informer cache.SharedIndexInformer
	notifier resource.Notifier
}

func newSecretService(informer cache.SharedIndexInformer, client corev1.CoreV1Interface) *secretService {
	notifier := resource.NewNotifier()
	informer.AddEventHandler(notifier)
	return &secretService{
		client:   client,
		informer: informer,
		notifier: notifier,
	}
}

func (svc secretService) Subscribe(listener resource.Listener) {
	svc.notifier.AddListener(listener)
}

func (svc secretService) Unsubscribe(listener resource.Listener) {
	svc.notifier.DeleteListener(listener)
}

func (svc secretService) Find(name, namespace string) (*v1.Secret, error) {
	key := fmt.Sprintf("%s/%s", namespace, name)
	item, exists, err := svc.informer.GetStore().GetByKey(key)
	if err != nil || !exists {
		return nil, err
	}

	secret, ok := item.(*v1.Secret)
	if !ok {
		return nil, fmt.Errorf("Incorrect item type: %T, should be: *Secret", item)
	}

	svc.ensureTypeMeta(secret)

	return secret, nil
}

func (svc secretService) List(namespace string, pagingParams pager.PagingParams) ([]*v1.Secret, error) {
	items, err := pager.FromIndexer(svc.informer.GetIndexer(), "namespace", namespace).Limit(pagingParams)
	if err != nil {
		return nil, err
	}

	var secrets []*v1.Secret
	for _, item := range items {
		secret, ok := item.(*v1.Secret)
		if !ok {
			return nil, fmt.Errorf("Incorrect item type: %T, should be: *Secret", item)
		}

		svc.ensureTypeMeta(secret)

		secrets = append(secrets, secret)
	}

	return secrets, nil
}

// Kubernetes API used by client-go doesn't provide kind and apiVersion so we have to add it here
// See: https://github.com/kubernetes/kubernetes/issues/3030
func (svc secretService) ensureTypeMeta(secret *v1.Secret) {
	secret.TypeMeta = svc.secretTypeMeta()
}

func (svc secretService) secretTypeMeta() metav1.TypeMeta {
	return metav1.TypeMeta{
		Kind:       "Secret",
		APIVersion: "v1",
	}
}
