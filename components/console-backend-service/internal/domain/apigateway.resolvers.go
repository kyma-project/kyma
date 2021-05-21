package domain

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/kyma-incubator/api-gateway/api/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *aPIRuleResolver) JSON(ctx context.Context, obj *v1alpha1.APIRule) (gqlschema.JSON, error) {
	return r.ag.JsonField(ctx, obj)
}

func (r *aPIRuleResolver) OwnerSubscription(ctx context.Context, obj *v1alpha1.APIRule) (*v1.OwnerReference, error) {
	return r.ag.GetOwnerSubscription(ctx, obj), nil
}

func (r *mutationResolver) CreateAPIRule(ctx context.Context, name string, namespace string, params v1alpha1.APIRuleSpec) (*v1alpha1.APIRule, error) {
	return r.ag.CreateAPIRule(ctx, name, namespace, params)
}

func (r *mutationResolver) UpdateAPIRule(ctx context.Context, name string, namespace string, generation int, params v1alpha1.APIRuleSpec) (*v1alpha1.APIRule, error) {
	return r.ag.UpdateAPIRule(ctx, name, namespace, int64(generation), params)
}

func (r *mutationResolver) DeleteAPIRule(ctx context.Context, name string, namespace string) (*v1alpha1.APIRule, error) {
	return r.ag.DeleteAPIRule(ctx, name, namespace)
}

func (r *queryResolver) APIRules(ctx context.Context, namespace string, serviceName *string, hostname *string) ([]*v1alpha1.APIRule, error) {
	return r.ag.APIRulesQuery(ctx, namespace, serviceName, hostname)
}

func (r *queryResolver) APIRule(ctx context.Context, name string, namespace string) (*v1alpha1.APIRule, error) {
	return r.ag.APIRuleQuery(ctx, name, namespace)
}

func (r *subscriptionResolver) APIRuleEvent(ctx context.Context, namespace string, serviceName *string) (<-chan *gqlschema.APIRuleEvent, error) {
	return r.ag.APIRuleEventSubscription(ctx, namespace, serviceName)
}

// APIRule returns gqlschema.APIRuleResolver implementation.
func (r *Resolver) APIRule() gqlschema.APIRuleResolver { return &aPIRuleResolver{r} }

type aPIRuleResolver struct{ *Resolver }
