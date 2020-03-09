package eventing

import (
	"context"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/eventing/disabled"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/module"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
)

type PluggableContainer struct {
	*module.Pluggable
	Resolver
	serviceFactory *resource.ServiceFactory
}

func New(serviceFactory *resource.ServiceFactory) (*PluggableContainer, error) {
	resolver := &PluggableContainer{
		Pluggable:      module.NewPluggable("eventing"),
		serviceFactory: serviceFactory,
	}

	err := resolver.Disable()
	if err != nil {
		return nil, err
	}

	return resolver, nil
}

func (r *PluggableContainer) Enable() error {
	triggerService := newTriggerService(r.serviceFactory)
	triggerConverter := newTriggerConverter()

	r.Pluggable.EnableAndSyncDynamicInformerFactory(r.serviceFactory.InformerFactory, func() {
		r.Resolver = &resolver{
			newTriggerResolver(triggerService, triggerConverter),
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

//go:generate failery -name=Resolver -case=underscore -output disabled -outpkg disabled
type Resolver interface {
	TriggersQuery(ctx context.Context, namespace string, subscriber *gqlschema.SubscriberInput) ([]gqlschema.Trigger, error)

	CreateTrigger(ctx context.Context, trigger gqlschema.TriggerCreateInput, ownerRef []gqlschema.OwnerReference) (*gqlschema.Trigger, error)
	CreateManyTriggers(ctx context.Context, triggers []gqlschema.TriggerCreateInput, ownerRef []gqlschema.OwnerReference) ([]gqlschema.Trigger, error)
	DeleteTrigger(ctx context.Context, trigger gqlschema.TriggerMetadata) (*gqlschema.Trigger, error)
	DeleteManyTriggers(ctx context.Context, triggers []gqlschema.TriggerMetadata) ([]gqlschema.Trigger, error)

	TriggerEventSubscription(ctx context.Context, namespace string, subscriber *gqlschema.SubscriberInput) (<-chan gqlschema.TriggerEvent, error)
}

type resolver struct {
	*triggerResolver
}
