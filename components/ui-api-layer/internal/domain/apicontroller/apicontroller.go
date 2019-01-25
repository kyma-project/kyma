package apicontroller

import (
	"context"
	"time"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/apicontroller/disabled"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/module"

	"github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma-project.io/clientset/versioned"
	"github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma-project.io/informers/externalversions"
	"github.com/pkg/errors"
	"k8s.io/client-go/rest"
)

type PluggableResolver struct {
	*module.Pluggable
	cfg *resolverConfig

	Resolver
	informerFactory externalversions.SharedInformerFactory
}

func New(restConfig *rest.Config, informerResyncPeriod time.Duration) (*PluggableResolver, error) {
	client, err := versioned.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing clientset")
	}

	resolver := &PluggableResolver{
		cfg: &resolverConfig{
			informerResyncPeriod: informerResyncPeriod,
			client:               client,
		},
		Pluggable: module.NewPluggable("apicontroller"),
	}
	err = resolver.Disable()

	return resolver, err
}

func (r *PluggableResolver) Enable() error {
	r.informerFactory = externalversions.NewSharedInformerFactory(r.cfg.client, r.cfg.informerResyncPeriod)
	apiService := newApiService(r.informerFactory.Gateway().V1alpha2().Apis().Informer())
	apiResolver, err := newApiResolver(apiService)
	if err != nil {
		return err
	}

	r.Pluggable.EnableAndSyncInformerFactory(r.informerFactory, func() {
		r.Resolver = &domainResolver{
			apiResolver: apiResolver,
		}
	})

	return nil
}

func (r *PluggableResolver) Disable() error {
	r.Pluggable.Disable(func(disabledErr error) {
		r.Resolver = disabled.NewResolver(disabledErr)
		r.informerFactory = nil
	})

	return nil
}

type resolverConfig struct {
	informerResyncPeriod time.Duration
	client               versioned.Interface
}

//go:generate failery -name=Resolver -case=underscore -output disabled -outpkg disabled
type Resolver interface {
	APIsQuery(ctx context.Context, environment *string, namespace *string, serviceName *string, hostname *string) ([]gqlschema.API, error)
}

type domainResolver struct {
	*apiResolver
}
