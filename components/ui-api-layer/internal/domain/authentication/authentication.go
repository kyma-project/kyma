package authentication

import (
	"context"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/authentication/disabled"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/module"
	"time"

	"github.com/kyma-project/kyma/components/idppreset/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/components/idppreset/pkg/client/informers/externalversions"
	"github.com/pkg/errors"
	"k8s.io/client-go/rest"
)


type PluggableResolver struct {
	*module.Pluggable
	cfg    *resolverConfig
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
			informerResyncPeriod:informerResyncPeriod,
			restConfig: restConfig,
			client: client,
		},
		Pluggable: module.NewPluggable("authentication"),
	}
	err = resolver.Disable()

	return resolver, err
}

func (r *PluggableResolver) Enable() error {
	r.informerFactory = externalversions.NewSharedInformerFactory(r.cfg.client, r.cfg.informerResyncPeriod)
	svc := newIDPPresetService(r.cfg.client.AuthenticationV1alpha1(), r.informerFactory.Authentication().V1alpha1().IDPPresets().Informer())

	r.Pluggable.Enable(r.informerFactory, func() {
		r.resolver = &domainResolver{
			idpPresetResolver: newIDPPresetResolver(svc),
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
	restConfig                *rest.Config
	informerResyncPeriod      time.Duration
	client *versioned.Clientset
}

//go:generate failery -name=resolver -case=underscore -output disabled -outpkg disabled
type resolver interface {
	IDPPresetsQuery(ctx context.Context, first *int, offset *int) ([]gqlschema.IDPPreset, error)
	IDPPresetQuery(ctx context.Context, name string) (*gqlschema.IDPPreset, error)
	DeleteIDPPresetMutation(ctx context.Context, name string) (*gqlschema.IDPPreset, error)
	CreateIDPPresetMutation(ctx context.Context, name string, issuer string, jwksURI string) (*gqlschema.IDPPreset, error)
}

type domainResolver struct {
	*idpPresetResolver
}
