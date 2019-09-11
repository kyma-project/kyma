package apicontroller

import (
	"context"
	"time"

	"github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma-project.io/v1alpha2"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/apicontroller/disabled"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/module"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/dynamic/dynamicinformer"
	"github.com/pkg/errors"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

type PluggableResolver struct {
	*module.Pluggable
	cfg *resolverConfig

	Resolver
	informerFactory dynamicinformer.DynamicSharedInformerFactory
}

var apisGroupVersionResource = schema.GroupVersionResource{
	Version:  v1alpha2.SchemeGroupVersion.Version,
	Group:    v1alpha2.SchemeGroupVersion.Group,
	Resource: "apis",
}

func New(restConfig *rest.Config, informerResyncPeriod time.Duration) (*PluggableResolver, error) {
	client, err := dynamic.NewForConfig(restConfig)
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
	r.informerFactory = dynamicinformer.NewDynamicSharedInformerFactory(r.cfg.client, r.cfg.informerResyncPeriod)
	aInformer := r.informerFactory.ForResource(apisGroupVersionResource).Informer()

	aResourceClient := r.cfg.client.Resource(apisGroupVersionResource)

	apiService := newApiService(aInformer, aResourceClient)
	apiResolver, err := newApiResolver(apiService)
	if err != nil {
		return err
	}

	r.Pluggable.EnableAndSyncCache(func(stopCh chan struct{}) {
		r.informerFactory.Start(stopCh)
		r.informerFactory.WaitForCacheSync(stopCh)

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
	client               dynamic.Interface
}

//go:generate failery -name=Resolver -case=underscore -output disabled -outpkg disabled
type Resolver interface {
	APIsQuery(ctx context.Context, namespace string, serviceName *string, hostname *string) ([]gqlschema.API, error)
	APIQuery(ctx context.Context, name string, namespace string) (*gqlschema.API, error)
	CreateAPI(ctx context.Context, name string, namespace string, params gqlschema.APIInput) (gqlschema.API, error)
	UpdateAPI(ctx context.Context, name string, namespace string, params gqlschema.APIInput) (gqlschema.API, error)
	DeleteAPI(ctx context.Context, name string, namespace string) (*gqlschema.API, error)
	APIEventSubscription(ctx context.Context, namespace string, serviceName *string) (<-chan gqlschema.ApiEvent, error)
}

type domainResolver struct {
	*apiResolver
}
