package apigateway

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-incubator/api-gateway/api/v1alpha1"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
)

type APIRuleList []*v1alpha1.APIRule

func (l *APIRuleList) Append() interface{} {
	e := &v1alpha1.APIRule{}
	*l = append(*l, e)
	return e
}

func (r *Resolver) APIRulesQuery(ctx context.Context, namespace string, serviceName *string, hostname *string) ([]*v1alpha1.APIRule, error) {
	items := APIRuleList{}
	var err error
	if serviceName != nil {
		err = r.Service().ListByIndex(apiRulesServiceIndex, apiRulesServiceIndexKey(namespace, serviceName), &items)
	} else if hostname != nil {
		err = r.Service().ListByIndex(apiRulesHostnameIndex, apiRulesHostnameIndexKey(namespace, hostname), &items)
	} else {
		err = r.Service().ListInNamespace(namespace, &items)
	}
	return items, err
}

func (r *Resolver) APIRuleQuery(ctx context.Context, name string, namespace string) (*v1alpha1.APIRule, error) {
	var result *v1alpha1.APIRule
	err := r.Service().GetInNamespace(name, namespace, &result)
	return result, err
}

func (r *Resolver) CreateAPIRule(ctx context.Context, name string, namespace string, params v1alpha1.APIRuleSpec) (*v1alpha1.APIRule, error) {
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
	var result *v1alpha1.APIRule
	err := r.Service().Create(apiRule, result)
	return result, err
}

func (r *Resolver) APIRuleEventSubscription(ctx context.Context, namespace string, serviceName *string) (<-chan *gqlschema.APIRuleEvent, error) {
	channel := make(chan *gqlschema.APIRuleEvent, 1)
	filter := func(apiRule v1alpha1.APIRule) bool {
		namespaceMatches := apiRule.Namespace == namespace
		serviceNameMatches := serviceName == nil || apiRule.Spec.Service.Name == serviceName
		return namespaceMatches && serviceNameMatches
	}

	unsubscribe, err := r.Service().Subscribe(NewEventHandler(channel, filter))
	if err != nil {
		return nil, err
	}

	go func() {
		defer close(channel)
		defer unsubscribe()
		<-ctx.Done()
	}()

	return channel, nil
}

func (r *Resolver) UpdateAPIRule(ctx context.Context, name string, namespace string, newSpec v1alpha1.APIRuleSpec) (*v1alpha1.APIRule, error) {
	var result *v1alpha1.APIRule
	err := r.Service().Update(name, namespace, result, func() error {
		result.Spec = newSpec
		return nil
	})
	return result, err
}

func (r *Resolver) DeleteAPIRule(ctx context.Context, name string, namespace string) (*v1alpha1.APIRule, error) {
	var result *v1alpha1.APIRule
	err := r.Service().DeleteInNamespace(namespace, name, result)
	return result, err
}
