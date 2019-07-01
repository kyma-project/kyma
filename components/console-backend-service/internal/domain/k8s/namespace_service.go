package k8s

import (
	"fmt"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"github.com/pkg/errors"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/application/pretty"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"
)

type namespaceService struct {
	informer cache.SharedIndexInformer
	client   corev1.CoreV1Interface
}

func newNamespaceService(informer cache.SharedIndexInformer, client corev1.CoreV1Interface) (*namespaceService, error) {
	return &namespaceService{
		informer: informer,
		client:   client,
	}, nil
}

func (svc *namespaceService) List() ([]*v1.Namespace, error) { //r error
	items := svc.informer.GetStore().List()

	var namespaces []*v1.Namespace
	for _, item := range items {
		namespace, ok := item.(*v1.Namespace)
		if !ok {
			return nil, fmt.Errorf("Incorrect item type: %T, should be: *Namespace", item)
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
		return nil, fmt.Errorf("Incorrect item type: %T, should be: *Namespace", item)
	}

	return namespace, nil
}

func (svc *namespaceService) Create(name string, labels gqlschema.Labels) (*v1.Namespace, error) {
	namespace := v1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "authentication.kyma-project.io/v1alpha1",
			Kind:       "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
	}

	return svc.client.Namespaces().Create(&namespace)
}

func (svc *namespaceService) Update(name string, labels gqlschema.Labels) (*v1.Namespace, error) {
	maxUpdateRetries  := 5
	var lastErr error
	for i := 0; i < maxUpdateRetries; i++ {
		namespace, err := svc.Find(name)

		if err != nil {
			return nil, errors.Wrapf(err, "while getting %s [%s]", pretty.Namespace, name)
		}
		if namespace == nil {
			return nil, apiErrors.NewNotFound(schema.GroupResource{
				Group:    "",
				Resource: "namespaces",
			}, name)
		}
		namespace.ObjectMeta.Labels = labels

		updated, err := svc.client.Namespaces().Update(namespace)
		switch {
		case err == nil:
			return updated, nil
		case apiErrors.IsConflict(err):
			lastErr = err
			continue
		default:
			return nil, errors.Wrapf(err, "while updating %s [%s]", pretty.Namespace, name)
		}
	}
	return nil, errors.Wrapf(lastErr, "couldn't update %s [%s], after %d retries", pretty.Namespace, name, maxUpdateRetries)
}

func (svc *namespaceService) Delete(name string) error {
	return svc.client.Namespaces().Delete(name, nil)
}
