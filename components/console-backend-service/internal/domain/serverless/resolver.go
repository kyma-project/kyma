package serverless

import (
	"context"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/serverless/disabled"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/serverless/function"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/module"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
)

type Container struct {
	Resolver
	*module.Pluggable
	serviceFactory *resource.ServiceFactory
}

type resolver struct {
	functionService FunctionService
}

//go:generate failery -name=Resolver -case=underscore -output disabled -outpkg disabled
type Resolver interface {
	FunctionsQuery(ctx context.Context, namespace string) ([]gqlschema.Function, error)
	FunctionQuery(ctx context.Context, name string, namespace string) (*gqlschema.Function, error)
	DeleteFunction(ctx context.Context, name string, namespace string) (*gqlschema.FunctionMutationOutput, error)
	CreateFunction(ctx context.Context, name string, namespace string, labels gqlschema.Labels, size string, runtime string) (*gqlschema.Function, error)
	UpdateFunction(ctx context.Context, name string, namespace string, params gqlschema.FunctionUpdateInput) (*gqlschema.Function, error)
}

func (r *Container) Enable() error {
	functionService := function.NewService(r.serviceFactory)

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

func New(serviceFactory *resource.ServiceFactory) (*Container, error) {
	container := &Container{
		Pluggable:      module.NewPluggable("serverless"),
		serviceFactory: serviceFactory,
	}

	err := container.Disable()
	if err != nil {
		return nil, err
	}

	return container, nil
}
