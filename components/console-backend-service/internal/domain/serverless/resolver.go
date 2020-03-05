package serverless

import (
	"context"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/serverless/disabled"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/serverless/function"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/shared"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/module"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
)

type PluggableContainer struct {
	*module.Pluggable
	cfg *resolverConfig

	Resolver
	serviceFactory *resource.ServiceFactory
}

func New(serviceFactory *resource.ServiceFactory, scaRetriever shared.ServiceCatalogAddonsRetriever) (*PluggableContainer, error) {
	container := &PluggableContainer{
		cfg: &resolverConfig{
			scaRetriever: scaRetriever,
		},
		Pluggable:      module.NewPluggable("serverless"),
		serviceFactory: serviceFactory,
	}

	err := container.Disable()
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (r *PluggableContainer) Enable() error {
	functionService := function.NewService(r.serviceFactory)

	r.Pluggable.EnableAndSyncDynamicInformerFactory(r.serviceFactory.InformerFactory, func() {
		r.Resolver = &resolver{
			functionService: functionService,
			scaRetriever:    r.cfg.scaRetriever,
		}
	})

	return nil
}

func (r *PluggableContainer) Disable() error {
	r.Pluggable.Disable(func(disabledErr error) {
		r.Resolver = disabled.NewResolver(disabledErr)
	})

	return nil
}

type resolverConfig struct {
	scaRetriever shared.ServiceCatalogAddonsRetriever
}

//go:generate failery -name=Resolver -case=underscore -output disabled -outpkg disabled
type Resolver interface {
	FunctionsQuery(ctx context.Context, namespace string) ([]gqlschema.Function, error)
	FunctionQuery(ctx context.Context, name string, namespace string) (*gqlschema.Function, error)
	ServiceBindingUsagesField(ctx context.Context, obj *gqlschema.Function) ([]gqlschema.ServiceBindingUsage, error)
	DeleteFunction(ctx context.Context, name string, namespace string) (*gqlschema.FunctionMutationOutput, error)
	CreateFunction(ctx context.Context, name string, namespace string, labels gqlschema.Labels, size string, runtime string) (*gqlschema.Function, error)
	UpdateFunction(ctx context.Context, name string, namespace string, params gqlschema.FunctionUpdateInput) (*gqlschema.Function, error)
}

type resolver struct {
	functionService FunctionService
	scaRetriever    shared.ServiceCatalogAddonsRetriever
}
