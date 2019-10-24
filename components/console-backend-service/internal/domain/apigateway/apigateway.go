package apigateway

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"k8s.io/client-go/rest"

	"github.com/kyma-incubator/api-gateway/api/v1alpha1"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/apigateway/disabled"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/module"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/dynamic/dynamicinformer"
	"k8s.io/client-go/dynamic"
)

type PluggableResolver struct {
	*module.Pluggable
	cfg *resolverConfig

	Resolver
	informerFactory dynamicinformer.DynamicSharedInformerFactory
}

var apiRulesGroupVersionResource = schema.GroupVersionResource{
	Version:  v1alpha1.GroupVersion.Version,
	Group:    v1alpha1.GroupVersion.Group,
	Resource: "apirules",
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
		Pluggable: module.NewPluggable("apigateway"),
	}
	err = resolver.Disable()

	return resolver, err
}

func (r *PluggableResolver) Enable() error {
	r.informerFactory = dynamicinformer.NewDynamicSharedInformerFactory(r.cfg.client, r.cfg.informerResyncPeriod)
	apiRuleInformer := r.informerFactory.ForResource(apiRulesGroupVersionResource).Informer()

	apiRuleResourceClient := r.cfg.client.Resource(apiRulesGroupVersionResource)

	apiRuleService := newApiRuleService(apiRuleInformer, apiRuleResourceClient)
	apiRuleResolver, err := newApiGatewayResolver(apiRuleService)
	if err != nil {
		return err
	}

	r.Pluggable.EnableAndSyncCache(func(stopCh chan struct{}) {
		r.informerFactory.Start(stopCh)
		r.informerFactory.WaitForCacheSync(stopCh)

		r.Resolver = &domainResolver{
			apiRuleResolver: apiRuleResolver,
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
	APIRulesQuery(ctx context.Context, namespace string, serviceName *string, hostname *string) ([]gqlschema.APIRule, error)
	APIRuleQuery(ctx context.Context, name string, namespace string) (*gqlschema.APIRule, error)
	CreateAPIRule(ctx context.Context, name string, namespace string, params gqlschema.APIRuleInput) (*gqlschema.APIRule, error)
	UpdateAPIRule(ctx context.Context, name string, namespace string, params gqlschema.APIRuleInput) (*gqlschema.APIRule, error)
	DeleteAPIRule(ctx context.Context, name string, namespace string) (*gqlschema.APIRule, error)
	APIRuleEventSubscription(ctx context.Context, namespace string, serviceName *string) (<-chan gqlschema.ApiRuleEvent, error)
}

type domainResolver struct {
	*apiRuleResolver
}
