package k8s

import (
	"fmt"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
)

type namespaceService struct {
	informer cache.SharedIndexInformer
	client   corev1.CoreV1Interface
}

func newNamespaceService(informer cache.SharedIndexInformer, client corev1.CoreV1Interface) (*namespaceService, error) {

	err := informer.AddIndexers(cache.Indexers{
		"envLabelSelector": func(obj interface{}) ([]string, error) {
			namespace, ok := obj.(*v1.Namespace)
			if !ok {
				return nil, fmt.Errorf("Incorrect item type: %T, should be: *Namespace", obj)
			}
			return []string{namespace.Labels["env"]}, nil
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "while adding indexers")
	}

	return &namespaceService{
		informer: informer,
		client:   client,
	}, nil
}

func (svc *namespaceService) List() ([]*v1.Namespace, error) {
	items, err := svc.informer.GetIndexer().ByIndex("envLabelSelector", "true")

	if err != nil {
		return nil, errors.Wrapf(err, "while listing %s", pretty.Namespace)
	}

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

func (svc *namespaceService) Delete(name string) error {
	return svc.client.Namespaces().Delete(name, nil)
}
