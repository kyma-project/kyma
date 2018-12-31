package kubeless

import (
	"context"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/kubeless/disabled"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/module"
	"time"

	"github.com/kubeless/kubeless/pkg/client/clientset/versioned"
	"github.com/kubeless/kubeless/pkg/client/informers/externalversions"
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
		return nil, errors.Wrap(err, "while initializing Clientset")
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
	functionService := newFunctionService(r.informerFactory.Kubeless().V1beta1().Functions().Informer())

	r.Pluggable.Enable(r.informerFactory, func() {
		r.resolver = &domainResolver{
			functionResolver: newFunctionResolver(functionService),
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
	FunctionsQuery(ctx context.Context, environment string, first *int, offset *int) ([]gqlschema.Function, error)
}

type domainResolver struct {
	*functionResolver
}