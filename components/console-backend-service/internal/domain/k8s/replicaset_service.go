package k8s

import (
	"context"
	"fmt"

	"github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"

	apps "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"

	appsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/pager"
	"k8s.io/client-go/tools/cache"
)

type replicaSetService struct {
	client   appsv1.AppsV1Interface
	informer cache.SharedIndexInformer
}

func newReplicaSetService(informer cache.SharedIndexInformer, client appsv1.AppsV1Interface) *replicaSetService {
	notifier := resource.NewNotifier()
	informer.AddEventHandler(notifier)
	return &replicaSetService{
		client:   client,
		informer: informer,
	}
}

func (svc *replicaSetService) Find(name, namespace string) (*apps.ReplicaSet, error) {
	key := fmt.Sprintf("%s/%s", namespace, name)
	item, exists, err := svc.informer.GetStore().GetByKey(key)
	if err != nil || !exists {
		return nil, err
	}

	replicaSet, ok := item.(*apps.ReplicaSet)
	if !ok {
		return nil, fmt.Errorf("Incorrect item type: %T, should be: *ReplicaSet", item)
	}

	svc.ensureTypeMeta(replicaSet)

	return replicaSet, nil
}

func (svc *replicaSetService) List(namespace string, pagingParams pager.PagingParams) ([]*apps.ReplicaSet, error) {
	items, err := pager.FromIndexer(svc.informer.GetIndexer(), "namespace", namespace).Limit(pagingParams)
	if err != nil {
		return nil, err
	}

	var replicaSets []*apps.ReplicaSet
	for _, item := range items {
		replicaSet, ok := item.(*apps.ReplicaSet)
		if !ok {
			return nil, fmt.Errorf("Incorrect item type: %T, should be: *ReplicaSet", item)
		}

		svc.ensureTypeMeta(replicaSet)

		replicaSets = append(replicaSets, replicaSet)
	}

	return replicaSets, nil
}

func (svc *replicaSetService) Update(name, namespace string, update apps.ReplicaSet) (*apps.ReplicaSet, error) {
	err := svc.checkUpdatePreconditions(name, namespace, update)
	if err != nil {
		return nil, err
	}

	updated, err := svc.client.ReplicaSets(namespace).Update(context.Background(), &update, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}

	svc.ensureTypeMeta(updated)

	return updated, nil
}

func (svc *replicaSetService) Delete(name, namespace string) error {
	return svc.client.ReplicaSets(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
}

func (svc *replicaSetService) checkUpdatePreconditions(name string, namespace string, update apps.ReplicaSet) error {
	errorList := field.ErrorList{}
	if name != update.Name {
		errorList = append(errorList, field.Invalid(field.NewPath("metadata.name"), update.Name, fmt.Sprintf("name of updated object does not match the original (%s)", name)))
	}
	if namespace != update.Namespace {
		errorList = append(errorList, field.Invalid(field.NewPath("metadata.namespace"), update.Namespace, fmt.Sprintf("namespace of updated object does not match the original (%s)", namespace)))
	}
	typeMeta := svc.replicaSetTypeMeta()
	if update.Kind != typeMeta.Kind {
		errorList = append(errorList, field.Invalid(field.NewPath("kind"), update.Kind, "ReplicaSet's kind should not be changed"))
	}
	if update.APIVersion != typeMeta.APIVersion {
		errorList = append(errorList, field.Invalid(field.NewPath("apiVersion"), update.APIVersion, "ReplicaSet's apiVersion should not be changed"))
	}

	if len(errorList) > 0 {
		return errors.NewInvalid(schema.GroupKind{
			Group: "",
			Kind:  "ReplicaSet",
		}, name, errorList)
	}

	return nil
}

// Kubernetes API used by client-go doesn't provide kind and apiVersion so we have to add it here
// See: https://github.com/kubernetes/kubernetes/issues/3030
func (svc *replicaSetService) ensureTypeMeta(replicaSet *apps.ReplicaSet) {
	replicaSet.TypeMeta = svc.replicaSetTypeMeta()
}

func (svc *replicaSetService) replicaSetTypeMeta() metav1.TypeMeta {
	return metav1.TypeMeta{
		Kind:       "ReplicaSet",
		APIVersion: "apps/v1",
	}
}
