package domain

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/ory/hydra-maester/api/v1alpha1"
)

func (r *mutationResolver) CreateOAuth2Client(ctx context.Context, name string, namespace string, params v1alpha1.OAuth2ClientSpec) (*v1alpha1.OAuth2Client, error) {
	return r.oauth.CreateOAuth2Client(ctx, name, namespace, params)
}

func (r *mutationResolver) UpdateOAuth2Client(ctx context.Context, name string, namespace string, generation int, params v1alpha1.OAuth2ClientSpec) (*v1alpha1.OAuth2Client, error) {
	return r.oauth.UpdateOAuth2Client(ctx, name, namespace, int64(generation), params)
}

func (r *mutationResolver) DeleteOAuth2Client(ctx context.Context, name string, namespace string) (*v1alpha1.OAuth2Client, error) {
	return r.oauth.DeleteOAuth2Client(ctx, name, namespace)
}

func (r *oAuth2ClientResolver) Error(ctx context.Context, obj *v1alpha1.OAuth2Client) (*v1alpha1.ReconciliationError, error) {
	return r.oauth.ErrorField(ctx, obj)
}

func (r *queryResolver) OAuth2Clients(ctx context.Context, namespace string) ([]*v1alpha1.OAuth2Client, error) {
	return r.oauth.OAuth2ClientsQuery(ctx, namespace)
}

func (r *queryResolver) OAuth2Client(ctx context.Context, name string, namespace string) (*v1alpha1.OAuth2Client, error) {
	return r.oauth.OAuth2ClientQuery(ctx, name, namespace)
}

func (r *subscriptionResolver) OAuth2ClientEvent(ctx context.Context, namespace string) (<-chan *gqlschema.OAuth2ClientEvent, error) {
	return r.oauth.OAuth2ClientSubscription(ctx, namespace)
}

// OAuth2Client returns gqlschema.OAuth2ClientResolver implementation.
func (r *Resolver) OAuth2Client() gqlschema.OAuth2ClientResolver { return &oAuth2ClientResolver{r} }

type oAuth2ClientResolver struct{ *Resolver }
