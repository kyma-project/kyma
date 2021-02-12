package apigateway

import (
	"context"
	"errors"
	"fmt"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"

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
	if serviceName != nil && hostname != nil {
		err = r.Service().ListByIndex(apiRulesServiceAndHostnameIndex, apiRulesServiceAndHostnameIndexKey(namespace, serviceName, hostname), &items)
	} else if hostname != nil {
		err = r.Service().ListByIndex(apiRulesHostnameIndex, apiRulesHostnameIndexKey(namespace, hostname), &items)
	} else if serviceName != nil {
		err = r.Service().ListByIndex(apiRulesServiceIndex, apiRulesServiceIndexKey(namespace, serviceName), &items)
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

func (r *Resolver) GetOwnerSubscription(ctx context.Context, rule *v1alpha1.APIRule) *metav1.OwnerReference {
	if rule.OwnerReferences == nil {
		return nil
	}

	for _, ownerRef := range rule.OwnerReferences {
		if ownerRef.Kind == "Subscription" {
			return &ownerRef
		}
	}

	return nil
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
	result := &v1alpha1.APIRule{}
	err := r.Service().Create(apiRule, result)
	return result, err
}

func (r *Resolver) APIRuleEventSubscription(ctx context.Context, namespace string, serviceName *string) (<-chan *gqlschema.APIRuleEvent, error) {
	channel := make(chan *gqlschema.APIRuleEvent, 1)
	filter := func(apiRule v1alpha1.APIRule) bool {
		namespaceMatches := apiRule.Namespace == namespace
		serviceNameMatches := serviceName == nil || *serviceName == *apiRule.Spec.Service.Name
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

func (r *Resolver) UpdateAPIRule(ctx context.Context, name string, namespace string, generation int64, newSpec v1alpha1.APIRuleSpec) (*v1alpha1.APIRule, error) {
	result := &v1alpha1.APIRule{}
	err := r.Service().UpdateInNamespace(name, namespace, result, func() error {
		if result.Generation > generation {
			return errors.New("resource already modified")
		}

		subscription := r.GetOwnerSubscription(ctx, result)
		if subscription != nil {
			return errors.New(fmt.Sprintf("API Rule is owned by %s", subscription.Name))
		}

		result.Spec = newSpec
		return nil
	})
	return result, err
}

func (r *Resolver) DeleteAPIRule(ctx context.Context, name string, namespace string) (*v1alpha1.APIRule, error) {
	result := &v1alpha1.APIRule{}
	err := r.Service().GetInNamespace(name, namespace, &result)
	if err != nil {
		return nil, err
	}
	subscription := r.GetOwnerSubscription(ctx, result)
	if subscription != nil {
		return nil, errors.New(fmt.Sprintf("API Rule is owned by %s", subscription.Name))
	}

	err = r.Service().DeleteInNamespace(namespace, name, result)
	return result, err
}

func (r *Resolver) JsonField(ctx context.Context, obj *v1alpha1.APIRule) (gqlschema.JSON, error) {
	return resource.ToJson(obj)
}
