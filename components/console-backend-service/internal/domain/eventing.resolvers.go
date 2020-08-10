package domain

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	v11 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "knative.dev/pkg/apis/duck/v1"
)

func (r *mutationResolver) CreateTrigger(ctx context.Context, namespace string, trigger gqlschema.TriggerCreateInput, ownerRef []*v11.OwnerReference) (*v1alpha1.Trigger, error) {
	return r.eventing.CreateTrigger(ctx, namespace, trigger, ownerRef)
}

func (r *mutationResolver) CreateManyTriggers(ctx context.Context, namespace string, triggers []*gqlschema.TriggerCreateInput, ownerRef []*v11.OwnerReference) ([]*v1alpha1.Trigger, error) {
	return r.eventing.CreateTriggers(ctx, namespace, triggers, ownerRef)
}

func (r *mutationResolver) DeleteTrigger(ctx context.Context, namespace string, triggerName string) (*v1alpha1.Trigger, error) {
	return r.eventing.DeleteTrigger(ctx, namespace, triggerName)
}

func (r *mutationResolver) DeleteManyTriggers(ctx context.Context, namespace string, triggerNames []string) ([]*v1alpha1.Trigger, error) {
	return r.eventing.DeleteTriggers(ctx, namespace, triggerNames)
}

func (r *queryResolver) Triggers(ctx context.Context, namespace string, subscriber *v1.Destination) ([]*v1alpha1.Trigger, error) {
	return r.eventing.TriggersQuery(ctx, namespace, subscriber)
}

func (r *subscriptionResolver) TriggerEvent(ctx context.Context, namespace string, subscriber *v1.Destination) (<-chan *gqlschema.TriggerEvent, error) {
	return r.eventing.TriggerEventSubscription(ctx, namespace, subscriber)
}

func (r *triggerResolver) Status(ctx context.Context, obj *v1alpha1.Trigger) (*gqlschema.TriggerStatus, error) {
	return r.eventing.StatusField(ctx, obj)
}

func (r *triggerSpecResolver) Filter(ctx context.Context, obj *v1alpha1.TriggerSpec) (gqlschema.JSON, error) {
	return r.eventing.FilterField(ctx, obj)
}

// Trigger returns gqlschema.TriggerResolver implementation.
func (r *Resolver) Trigger() gqlschema.TriggerResolver { return &triggerResolver{r} }

// TriggerSpec returns gqlschema.TriggerSpecResolver implementation.
func (r *Resolver) TriggerSpec() gqlschema.TriggerSpecResolver { return &triggerSpecResolver{r} }

type triggerResolver struct{ *Resolver }
type triggerSpecResolver struct{ *Resolver }
