package apigateway

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-incubator/api-gateway/api/v1alpha1"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
	"github.com/pkg/errors"
)

//go:generate mockery -name=apiRuleSvc -output=automock -outpkg=automock -case=underscore
type apiRuleSvc interface {
	List(namespace string, serviceName *string, hostname *string) ([]*v1alpha1.APIRule, error)
	Find(name string, namespace string) (*v1alpha1.APIRule, error)
	Create(api *v1alpha1.APIRule) (*v1alpha1.APIRule, error)
	Update(name, namespace string, api v1alpha1.APIRuleSpec) (*v1alpha1.APIRule, error)
	Delete(name string, namespace string) (*v1alpha1.APIRule, error)
	Subscribe(listener resource.EventHandlerProvider) resource.Unsubscribe
}

type apiRuleResolver struct {
	apiRuleSvc apiRuleSvc
}

func newApiRuleResolver(svc apiRuleSvc) (*apiRuleResolver, error) {
	if svc == nil {
		return nil, errors.New("Nil pointer for apiRuleSvc")
	}

	return &apiRuleResolver{
		apiRuleSvc: svc,
	}, nil
}

func (ar *apiRuleResolver) APIRulesQuery(ctx context.Context, namespace string, serviceName *string, hostname *string) ([]*v1alpha1.APIRule, error) {
	return ar.apiRuleSvc.List(namespace, serviceName, hostname)
}

func (ar *apiRuleResolver) APIRuleQuery(ctx context.Context, name string, namespace string) (*v1alpha1.APIRule, error) {
	return ar.apiRuleSvc.Find(name, namespace)
}

func (ar *apiRuleResolver) CreateAPIRule(ctx context.Context, name string, namespace string, params v1alpha1.APIRuleSpec) (*v1alpha1.APIRule, error) {
	apiRule := &v1alpha1.APIRule{
		TypeMeta: metav1.TypeMeta{
			APIVersion: apiRulesGroupVersionResource.GroupVersion().String(),
			Kind:       apiRulesKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: params,
	}
	return ar.apiRuleSvc.Create(apiRule)
}

func (ar *apiRuleResolver) APIRuleEventSubscription(ctx context.Context, namespace string, serviceName *string) (<-chan *gqlschema.APIRuleEvent, error) {
	channel := make(chan *gqlschema.APIRuleEvent, 1)
	filter := func(apiRule v1alpha1.APIRule) bool {
		namespaceMatches := apiRule.Namespace == namespace
		serviceNameMatches := serviceName == nil || apiRule.Spec.Service.Name == serviceName
		return namespaceMatches && serviceNameMatches
	}

	unsubscribe := ar.apiRuleSvc.Subscribe(NewEventHandler(channel, filter))
	go func() {
		defer close(channel)
		defer unsubscribe()
		<-ctx.Done()
	}()

	return channel, nil
}

func (ar *apiRuleResolver) UpdateAPIRule(ctx context.Context, name string, namespace string, params v1alpha1.APIRuleSpec) (*v1alpha1.APIRule, error) {
	return ar.apiRuleSvc.Update(name, namespace, params)
}

func (ar *apiRuleResolver) DeleteAPIRule(ctx context.Context, name string, namespace string) (*v1alpha1.APIRule, error) {
	return ar.apiRuleSvc.Delete(name, namespace)
}
