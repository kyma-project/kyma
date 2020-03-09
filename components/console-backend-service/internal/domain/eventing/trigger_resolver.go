package eventing

import (
	"context"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"

	"github.com/golang/glog"
	"github.com/pkg/errors"
)

type triggerResolver struct {
	service   triggerSvc
	converter gqlTriggerConverter
}

func newTriggerResolver(svc triggerSvc, converter gqlTriggerConverter) *triggerResolver {
	return &triggerResolver{
		service:   svc,
		converter: converter,
	}
}

func (r *triggerResolver) TriggersQuery(ctx context.Context, namespace string, subscriber *gqlschema.SubscriberInput) ([]gqlschema.Trigger, error) {
	if subscriber == nil {
		glog.Error(errors.Errorf("%s cannot be empty in order to resolve 'trigger' field", Trigger))
		return nil, gqlerror.NewInternal()
	}



	return nil, nil
}

func (r *triggerResolver) CreateTrigger(ctx context.Context, trigger gqlschema.TriggerCreateInput, ownerRef []gqlschema.OwnerReference) (*gqlschema.Trigger, error) {
	return nil, nil
}

func (r *triggerResolver) CreateManyTriggers(ctx context.Context, triggers []gqlschema.TriggerCreateInput, ownerRef []gqlschema.OwnerReference) ([]gqlschema.Trigger, error) {
	return nil, nil
}

func (r *triggerResolver) DeleteTrigger(ctx context.Context, trigger gqlschema.TriggerMetadata) (*gqlschema.Trigger, error) {
	return nil, nil
}

func (r *triggerResolver) DeleteManyTriggers(ctx context.Context, triggers []gqlschema.TriggerMetadata) ([]gqlschema.Trigger, error) {
	return nil, nil
}

func (r *triggerResolver) TriggerEventSubscription(ctx context.Context, namespace string, subscriber *gqlschema.SubscriberInput) (<-chan gqlschema.TriggerEvent, error) {
	return nil, nil
}
