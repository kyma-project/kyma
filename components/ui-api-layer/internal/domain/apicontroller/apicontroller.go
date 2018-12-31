package apicontroller

import (
	"context"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/apicontroller/disabled"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/module"
	"time"

	"github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma-project.io/clientset/versioned"
	"github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma-project.io/informers/externalversions"
	"github.com/pkg/errors"
	"k8s.io/client-go/rest"
)

type PluggableResolver struct {
	*module.Pluggable
	cfg *resolverConfig
	resolver

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
			restConfig:           restConfig,
			client:               client,
		},
		Pluggable: module.NewPluggable("authentication"),
	}
	err = resolver.Disable()

	return resolver, err
}

func (r *PluggableResolver) Enable() error {
	r.informerFactory = externalversions.NewSharedInformerFactory(r.cfg.client, r.cfg.informerResyncPeriod)
	apiService := newApiService(r.informerFactory.Gateway().V1alpha2().Apis().Informer())

	r.Pluggable.Enable(r.informerFactory, func() {
		r.resolver = &domainResolver{
			apiResolver:     newApiResolver(apiService),
		}
	})

	return nil
}

func (r *PluggableResolver) Disable() error {
	r.Pluggable.Disable(func(disabledErr error) {
		r.resolver = disabled.NewResolver(disabledErr)
	})

	return nil
}

type resolverConfig struct {
	restConfig           *rest.Config
	informerResyncPeriod time.Duration
	client               *versioned.Clientset
}

//go:generate failery -name=resolver -case=underscore -output disabled -outpkg disabled
type resolver interface {
	APIsQuery(ctx context.Context, environment string, serviceName *string, hostname *string) ([]gqlschema.API, error)
}

type domainResolver struct {
	*apiResolver
}
