package apigateway

import (
	"context"

	"github.com/kyma-incubator/api-gateway/api/v1alpha1"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/apigateway/listener"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/apigateway/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"
	"github.com/pkg/errors"
)

//go:generate mockery -name=apiRuleSvc -output=automock -outpkg=automock -case=underscore
type apiRuleSvc interface {
	List(namespace string, serviceName *string, hostname *string) ([]*v1alpha1.APIRule, error)
	Find(name string, namespace string) (*v1alpha1.APIRule, error)
	Create(api *v1alpha1.APIRule) (*v1alpha1.APIRule, error)
	Update(api *v1alpha1.APIRule) (*v1alpha1.APIRule, error)
	Delete(name string, namespace string) error
	Subscribe(listener resource.Listener)
	Unsubscribe(listener resource.Listener)
}

//go:generate mockery -name=apiRuleConv -output=automock -outpkg=automock -case=underscore
type apiRuleConv interface {
	ToApiRule(name string, namespace string, in gqlschema.APIRuleInput) *v1alpha1.APIRule
}

type apiRuleResolver struct {
	apiRuleSvc apiRuleSvc
	apiRuleCon apiRuleConv
}

func newApiRuleResolver(svc apiRuleSvc) (*apiRuleResolver, error) {
	if svc == nil {
		return nil, errors.New("Nil pointer for apiRuleSvc")
	}

	return &apiRuleResolver{
		apiRuleSvc: svc,
		apiRuleCon: &apiRuleConverter{},
	}, nil
}

func (ar *apiRuleResolver) APIRulesQuery(ctx context.Context, namespace string, serviceName *string, hostname *string) ([]*v1alpha1.APIRule, error) {
	return ar.apiRuleSvc.List(namespace, serviceName, hostname)
}

func (ar *apiRuleResolver) APIRuleQuery(ctx context.Context, name string, namespace string) (*v1alpha1.APIRule, error) {
	return ar.apiRuleSvc.Find(name, namespace)
}

func (ar *apiRuleResolver) CreateAPIRule(ctx context.Context, name string, namespace string, params gqlschema.APIRuleInput) (*v1alpha1.APIRule, error) {
	apiRuleObject := ar.apiRuleCon.ToApiRule(name, namespace, params)

	return ar.apiRuleSvc.Create(apiRuleObject)
}

func (ar *apiRuleResolver) APIRuleEventSubscription(ctx context.Context, namespace string, serviceName *string) (<-chan *gqlschema.APIRuleEvent, error) {
	channel := make(chan *gqlschema.APIRuleEvent, 1)
	filter := func(apiRule *v1alpha1.APIRule) bool {
		if serviceName == nil {
			return apiRule != nil && apiRule.Namespace == namespace
		}
		return apiRule != nil && apiRule.Namespace == namespace && apiRule.Spec.Service.Name == serviceName
	}

	apiRuleListener := listener.NewApiRule(channel, filter, ApiRuleUnstructuredExtractor{})

	ar.apiRuleSvc.Subscribe(apiRuleListener)
	go func() {
		defer close(channel)
		defer ar.apiRuleSvc.Unsubscribe(apiRuleListener)
		<-ctx.Done()
	}()

	return channel, nil
}

func (ar *apiRuleResolver) UpdateAPIRule(ctx context.Context, name string, namespace string, params gqlschema.APIRuleInput) (*v1alpha1.APIRule, error) {
	apiRuleObject := ar.apiRuleCon.ToApiRule(name, namespace, params)

	return ar.apiRuleSvc.Update(apiRuleObject)
}

func (ar *apiRuleResolver) DeleteAPIRule(ctx context.Context, name string, namespace string) (*v1alpha1.APIRule, error) {
	apiRuleObj, err := ar.apiRuleSvc.Find(name, namespace)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting %s `%s` in namespace `%s`", pretty.APIRule, name, namespace))
		return nil, gqlerror.New(err, pretty.APIRules, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	apiRuleCopy := apiRuleObj.DeepCopy()
	err = ar.apiRuleSvc.Delete(name, namespace)
	return apiRuleCopy, err
}
