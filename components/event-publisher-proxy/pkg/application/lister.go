package application

import (
	"context"
	"errors"
	"time"

	applicationv1alpha1 "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	kymalogger "github.com/kyma-project/kyma/components/eventing-controller/logger"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"

	"github.com/kyma-project/kyma/components/event-publisher-proxy/pkg/informers"
)

type Lister struct {
	lister cache.GenericLister
}

func NewLister(ctx context.Context, client dynamic.Interface) *Lister {
	const defaultResync = 10 * time.Second
	gvr := GroupVersionResource()
	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(client, defaultResync, v1.NamespaceAll, nil)
	factory.ForResource(gvr)
	lister := factory.ForResource(gvr).Lister()
	logger, _ := kymalogger.New("json", "error")
	informers.WaitForCacheSyncOrDie(ctx, factory, logger)
	return &Lister{lister: lister}
}

func (l Lister) Get(name string) (*applicationv1alpha1.Application, error) {
	object, err := l.lister.Get(name)
	if err != nil {
		return nil, err
	}

	u, ok := object.(*unstructured.Unstructured)
	if !ok {
		return nil, errors.New("failed to convert runtime object to unstructured")
	}

	a := &applicationv1alpha1.Application{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, a); err != nil {
		return nil, err
	}

	return a, nil
}

func GroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    applicationv1alpha1.SchemeGroupVersion.Group,
		Version:  applicationv1alpha1.SchemeGroupVersion.Version,
		Resource: "applications",
	}
}
