package domain

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
)

func (r *mutationResolver) CreateSubscription(ctx context.Context, name string, namespace string, params gqlschema.EventSubscriptionSpecInput) (*v1alpha1.Subscription, error) {
	return r.bebEventing.CreateEventSubscription(ctx, namespace, name, params)
}

func (r *mutationResolver) UpdateSubscription(ctx context.Context, name string, namespace string, params gqlschema.EventSubscriptionSpecInput) (*v1alpha1.Subscription, error) {
	return r.bebEventing.UpdateEventSubscription(ctx, namespace, name, params)
}

func (r *mutationResolver) DeleteSubscription(ctx context.Context, name string, namespace string) (*v1alpha1.Subscription, error) {
	return r.bebEventing.DeleteEventSubscription(ctx, namespace, name)
}

func (r *queryResolver) EventSubscriptions(ctx context.Context, ownerName string, namespace string) ([]*v1alpha1.Subscription, error) {
	return r.bebEventing.EventSubscriptionsQuery(ctx, ownerName, namespace)
}

func (r *subscriptionResolver) SubscriptionSubscription(ctx context.Context, ownerName string, namespace string) (<-chan *gqlschema.SubscriptionEvent, error) {
	return r.bebEventing.SubscribeEventSubscription(ctx, ownerName, namespace)
}
