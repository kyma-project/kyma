package serverless

import (
	"context"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/shared"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/serverless/disabled"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/module"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
)

type Config struct {
	UsageKind string `envconfig:"default=function"`
}

type PluggableContainer struct {
	*module.Pluggable
	cfg *resolverConfig

	Resolver
	serviceFactory *resource.ServiceFactory
}

func New(serviceFactory *resource.ServiceFactory, cfg Config, scaRetriever shared.ServiceCatalogAddonsRetriever) (*PluggableContainer, error) {
	resolver := &PluggableContainer{
		Pluggable: module.NewPluggable("serverless"),
		cfg: &resolverConfig{
			cfg:          &cfg,
			scaRetriever: scaRetriever,
		},
		serviceFactory: serviceFactory,
	}

	err := resolver.Disable()
	if err != nil {
		return nil, err
	}

	return resolver, nil
}

func (r *PluggableContainer) Enable() error {
	functionService := newFunctionService(r.serviceFactory)
	functionConverter := newFunctionConverter()

	r.Pluggable.EnableAndSyncDynamicInformerFactory(r.serviceFactory.InformerFactory, func() {
		r.Resolver = &domainResolver{
			functionResolver: newFunctionResolver(functionService, functionConverter, r.cfg.cfg, r.cfg.scaRetriever),
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
	cfg          *Config
	scaRetriever shared.ServiceCatalogAddonsRetriever
}

//go:generate failery -name=Resolver -case=underscore -output disabled -outpkg disabled
type Resolver interface {
	FunctionQuery(ctx context.Context, name string, namespace string) (*gqlschema.Function, error)
	FunctionsQuery(ctx context.Context, namespace string) ([]gqlschema.Function, error)

	CreateFunction(ctx context.Context, name string, namespace string, params gqlschema.FunctionMutationInput) (*gqlschema.Function, error)
	UpdateFunction(ctx context.Context, name string, namespace string, params gqlschema.FunctionMutationInput) (*gqlschema.Function, error)
	DeleteFunction(ctx context.Context, function gqlschema.FunctionMetadataInput) (*gqlschema.FunctionMetadata, error)
	DeleteManyFunctions(ctx context.Context, functions []gqlschema.FunctionMetadataInput) ([]gqlschema.FunctionMetadata, error)

	FunctionEventSubscription(ctx context.Context, namespace string, functionName *string) (<-chan gqlschema.FunctionEvent, error)
}

type domainResolver struct {
	*functionResolver
}
