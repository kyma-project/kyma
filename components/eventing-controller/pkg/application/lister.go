package application

import (
	"context"
	"errors"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"

	applicationv1alpha1 "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/informers"
)

type Lister struct {
	lister cache.GenericLister
}

func NewLister(ctx context.Context, client dynamic.Interface) *Lister {
	gvr := GroupVersionResource()
	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(client, 10*time.Second, v1.NamespaceAll, nil)
	factory.ForResource(gvr)
	lister := factory.ForResource(gvr).Lister()
	informers.WaitForCacheSyncOrDie(ctx, factory)
	return &Lister{lister: lister}
}

func (l Lister) Get(name string) (*applicationv1alpha1.Application, error) {
	object, err := l.lister.Get(name)
	if err != nil {
		return nil, err
	}

	applicationUnstructured, ok := object.(*unstructured.Unstructured)
	if !ok {
		return nil, errors.New("failed to convert runtime object to unstructured")
	}

	application := &applicationv1alpha1.Application{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(applicationUnstructured.Object, application); err != nil {
		return nil, err
	}

	return application, nil
}

func GroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    applicationv1alpha1.SchemeGroupVersion.Group,
		Version:  applicationv1alpha1.SchemeGroupVersion.Version,
		Resource: "applications",
	}
}

func GroupVersionKind() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   applicationv1alpha1.SchemeGroupVersion.Group,
		Version: applicationv1alpha1.SchemeGroupVersion.Version,
		Kind:    "Application",
	}
}
