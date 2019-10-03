package serverless

import (
	"context"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/serverless/disabled"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/module"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/dynamic/dynamicinformer"
	"github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"time"
)

type Container struct {
	Resolver
	*module.Pluggable
	informerResyncPeriod time.Duration
	dynamicClient dynamic.Interface
	informerFactory dynamicinformer.DynamicSharedInformerFactory
}

type resolver struct {
	container *Container
	functionService *functionService
}

//go:generate failery -name=Resolver -case=underscore -output disabled -outpkg disabled
type Resolver interface {
	FunctionsQuery(ctx context.Context, namespace string) ([]gqlschema.Function, error)
	DeleteFunction(ctx context.Context, name string, namespace string) (gqlschema.FunctionMutationOutput, error)
}

func (r *Container) Enable() error {
	informerFactory := dynamicinformer.NewDynamicSharedInformerFactory(r.dynamicClient, r.informerResyncPeriod)
	r.informerFactory = informerFactory

	functionClient := r.dynamicClient.Resource(schema.GroupVersionResource{
		Version: v1alpha1.SchemeGroupVersion.Version,
		Group: v1alpha1.SchemeGroupVersion.Group,
		Resource: "functions",
	})

	functionService := newFunctionService(informerFactory.ForResource(schema.GroupVersionResource{
		Version: v1alpha1.SchemeGroupVersion.Version,
		Group: v1alpha1.SchemeGroupVersion.Group,
		Resource: "functions",
	}).Informer(), functionClient)

	r.Pluggable.EnableAndSyncDynamicInformerFactory(r.informerFactory, func() {
		r.Resolver = &resolver{
			container: r,
			functionService: functionService}
	})

	return nil
}

func (r *Container) Disable() error {
	r.Pluggable.Disable(func(disabledErr error) {
		r.Resolver = disabled.NewResolver(disabledErr)
		r.informerFactory = nil
	})

	return nil
}

func New(config *rest.Config, informerResyncPeriod time.Duration) (*Container, error) {
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing Dynamic Clientset")
	}

	container := &Container{
		Pluggable: module.NewPluggable("serverless"),
		informerResyncPeriod: informerResyncPeriod,
		dynamicClient: dynamicClient,
	}

	err = container.Disable()
	if err != nil {
		return nil, err
	}

	return container, nil
}