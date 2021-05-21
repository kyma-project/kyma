package k8s

import (
	"context"
	"fmt"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/application/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/retry"
)

type namespaceService struct {
	informer cache.SharedIndexInformer
	podsSvc  podSvc
	client   corev1.CoreV1Interface
	notifier resource.Notifier
}

func newNamespaceService(informer cache.SharedIndexInformer, podsSvc podSvc, client corev1.CoreV1Interface) (*namespaceService, error) {
	notifier := resource.NewNotifier()
	informer.AddEventHandler(notifier)
	return &namespaceService{
		informer: informer,
		podsSvc:  podsSvc,
		client:   client,
		notifier: notifier,
	}, nil
}

func (svc *namespaceService) List() ([]*v1.Namespace, error) {
	items := svc.informer.GetStore().List()

	var namespaces []*v1.Namespace
	for _, item := range items {
		namespace, ok := item.(*v1.Namespace)
		if !ok {
			return nil, fmt.Errorf("incorrect item type: %T, should be: *Namespace", item)
		}
		namespaces = append(namespaces, namespace)
	}

	return namespaces, nil
}

func (svc *namespaceService) Find(name string) (*v1.Namespace, error) {
	item, exists, err := svc.informer.GetStore().GetByKey(name)

	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, nil
	}

	namespace, ok := item.(*v1.Namespace)
	if !ok {
		return nil, fmt.Errorf("incorrect item type: %T, should be: *Namespace", item)
	}

	return namespace, nil
}

func (svc *namespaceService) Create(name string, labels gqlschema.Labels) (*v1.Namespace, error) {
	namespace := v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
	}

	return svc.client.Namespaces().Create(context.Background(), &namespace, metav1.CreateOptions{})
}

func (svc *namespaceService) Update(name string, labels gqlschema.Labels) (*v1.Namespace, error) {
	var updated *v1.Namespace
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		namespace, err := svc.Find(name)

		if err != nil {
			return errors.Wrapf(err, "while getting %s [%s]", pretty.Namespace, name)
		}
		if namespace == nil {
			return apiErrors.NewNotFound(schema.GroupResource{
				Group:    "",
				Resource: "namespaces",
			}, name)
		}
		namespace.ObjectMeta.Labels = labels

		updated, err = svc.client.Namespaces().Update(context.Background(), namespace, metav1.UpdateOptions{})

		return errors.Wrapf(err, "while updating %s [%s]", pretty.Namespace, name)
	})

	if err != nil {
		return nil, errors.Wrapf(err, "couldn't update %s [%s], after %d retries", pretty.Namespace, name, retry.DefaultRetry.Steps)
	}
	return updated, nil
}

func (svc *namespaceService) Delete(name string) error {
	return svc.client.Namespaces().Delete(context.Background(), name, metav1.DeleteOptions{})
}

func (svc *namespaceService) Subscribe(listener resource.Listener) {
	svc.notifier.AddListener(listener)
}

func (svc *namespaceService) Unsubscribe(listener resource.Listener) {
	svc.notifier.DeleteListener(listener)
}
