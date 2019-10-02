package serverless

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/serverless/disabled"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/module"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/dynamic/dynamicinformer"
	"github.com/pkg/errors"
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
}

//go:generate failery -name=Resolver -case=underscore -output disabled -outpkg disabled
type Resolver interface {

}

func (r *Container) Enable() error {
	informerFactory := dynamicinformer.NewDynamicSharedInformerFactory(r.dynamicClient, r.informerResyncPeriod)
	r.informerFactory = informerFactory

	r.Pluggable.EnableAndSyncDynamicInformerFactory(r.informerFactory, func() {
		r.Resolver = &resolver{
			container: r,
		}
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
		Pluggable: module.NewPluggable("serverless"), // TODO change to serverless
		informerResyncPeriod: informerResyncPeriod,
		dynamicClient: dynamicClient,
	}

	err = container.Disable()
	if err != nil {
		return nil, err
	}

	return container, nil
}