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
	resolver
	cfg    *resolverConfig
	stopCh    chan struct{}
	isEnabled bool

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
	}
	err = resolver.Disable()

	return resolver, err
}

func (r *PluggableResolver) Name() string {
	return "authentication"
}

func (r *PluggableResolver) IsEnabled() bool {
	return r.isEnabled
}

func (r *PluggableResolver) Enable() error {
	r.isEnabled = true
	r.stopCh = make(chan struct{})

	r.informerFactory = externalversions.NewSharedInformerFactory(r.cfg.client, r.cfg.informerResyncPeriod)
	svc := newIDPPresetService(r.cfg.client.AuthenticationV1alpha1(), r.informerFactory.Authentication().V1alpha1().IDPPresets().Informer())

	go func() {
		r.startAndWaitForCacheSync()
		r.resolver = &domainResolver{
			idpPresetResolver: newIDPPresetResolver(svc),
		}
	}()

	return nil
}

func (r *PluggableResolver) Disable() error {
	r.isEnabled = false

	if r.stopCh != nil {
		close(r.stopCh)
	}

	disabledErr := module.DisabledError(r)
	r.resolver = disabled.NewResolver(disabledErr)

	return nil
}

func (r *PluggableResolver) StopCacheSyncOnClose(stopCh <-chan struct{}) {
	go func() {
		<-stopCh
		close(r.stopCh)
	}()
}

func (r *PluggableResolver) startAndWaitForCacheSync() {
	r.informerFactory.Start(r.stopCh)
	r.informerFactory.WaitForCacheSync(r.stopCh)
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
