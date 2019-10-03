package serverless

import (
	"context"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/serverless/disabled"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/module"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
	"github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"time"
)

type Container struct {
	Resolver
	*module.Pluggable
	serviceFactory *resource.ServiceFactory
}

type resolver struct {
	functionService *functionService
}

//go:generate failery -name=Resolver -case=underscore -output disabled -outpkg disabled
type Resolver interface {
	FunctionsQuery(ctx context.Context, namespace string) ([]gqlschema.Function, error)
	DeleteFunction(ctx context.Context, name string, namespace string) (gqlschema.FunctionMutationOutput, error)
}

func (r *Container) Enable() error {
	functionService := newFunctionService(r.serviceFactory.ForResource(schema.GroupVersionResource{
		Version:  v1alpha1.SchemeGroupVersion.Version,
		Group:    v1alpha1.SchemeGroupVersion.Group,
		Resource: "functions",
	}))

	r.Pluggable.EnableAndSyncDynamicInformerFactory(r.serviceFactory.InformerFactory, func() {
		r.Resolver = &resolver{
			functionService: functionService,
		}
	})

	return nil
}

func (r *Container) Disable() error {
	r.Pluggable.Disable(func(disabledErr error) {
		r.Resolver = disabled.NewResolver(disabledErr)
	})

	return nil
}

func New(config *rest.Config, informerResyncPeriod time.Duration) (*Container, error) {
	serviceFactory, err := resource.NewServiceFactory(config, informerResyncPeriod)
	container := &Container{
		Pluggable:      module.NewPluggable("serverless"),
		serviceFactory: serviceFactory,
	}

	err = container.Disable()
	if err != nil {
		return nil, err
	}

	return container, nil
}