package domain

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/eventing/pkg/apis/eventing/v1alpha1"
)

func (r *mutationResolver) CreateTrigger(ctx context.Context, namespace string, trigger gqlschema.TriggerCreateInput, ownerRef []*v1.OwnerReference) (*v1alpha1.Trigger, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CreateManyTriggers(ctx context.Context, namespace string, triggers []*gqlschema.TriggerCreateInput, ownerRef []*v1.OwnerReference) ([]*v1alpha1.Trigger, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) DeleteTrigger(ctx context.Context, namespace string, triggerName string) (*v1alpha1.Trigger, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) DeleteManyTriggers(ctx context.Context, namespace string, triggerNames []string) ([]*v1alpha1.Trigger, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Triggers(ctx context.Context, namespace string, serviceName string) ([]*v1alpha1.Trigger, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *subscriptionResolver) TriggerEvent(ctx context.Context, namespace string, serviceName string) (<-chan *gqlschema.TriggerEvent, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *triggerResolver) Status(ctx context.Context, obj *v1alpha1.Trigger) (*gqlschema.TriggerStatus, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *triggerSpecResolver) Filter(ctx context.Context, obj *v1alpha1.TriggerSpec) (gqlschema.JSON, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *triggerSpecResolver) Port(ctx context.Context, obj *v1alpha1.TriggerSpec) (uint32, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *triggerSpecResolver) Path(ctx context.Context, obj *v1alpha1.TriggerSpec) (string, error) {
	panic(fmt.Errorf("not implemented"))
}

// Trigger returns gqlschema.TriggerResolver implementation.
func (r *Resolver) Trigger() gqlschema.TriggerResolver { return &triggerResolver{r} }

// TriggerSpec returns gqlschema.TriggerSpecResolver implementation.
func (r *Resolver) TriggerSpec() gqlschema.TriggerSpecResolver { return &triggerSpecResolver{r} }

type triggerResolver struct{ *Resolver }
type triggerSpecResolver struct{ *Resolver }
