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
	ToGQL(in *v1alpha1.APIRule) (*gqlschema.APIRule, error)
	ToGQLs(in []*v1alpha1.APIRule) ([]gqlschema.APIRule, error)
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

func (ar *apiRuleResolver) APIRulesQuery(ctx context.Context, namespace string, serviceName *string, hostname *string) ([]gqlschema.APIRule, error) {
	apiRulessObj, err := ar.apiRuleSvc.List(namespace, serviceName, hostname)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing %s for service name %v, hostname %v", pretty.APIRules, serviceName, hostname))
		return nil, gqlerror.New(err, pretty.APIRules, gqlerror.WithNamespace(namespace))
	}
	apis, err := ar.apiRuleCon.ToGQLs(apiRulessObj)
	if err != nil {
		return nil, err
	}
	return apis, nil
}

func (ar *apiRuleResolver) APIRuleQuery(ctx context.Context, name string, namespace string) (*gqlschema.APIRule, error) {
	apiRuleObj, err := ar.apiRuleSvc.Find(name, namespace)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting %s `%s` in namespace `%s`", pretty.APIRule, name, namespace))
		return nil, gqlerror.New(err, pretty.APIRules, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	if apiRuleObj == nil {
		return nil, nil
	}

	apiRule, err := ar.apiRuleCon.ToGQL(apiRuleObj)
	if err != nil {
		return nil, err
	}
	return apiRule, nil
}

func (ar *apiRuleResolver) CreateAPIRule(ctx context.Context, name string, namespace string, params gqlschema.APIRuleInput) (*gqlschema.APIRule, error) {
	apiRuleObject := ar.apiRuleCon.ToApiRule(name, namespace, params)

	apiRule, err := ar.apiRuleSvc.Create(apiRuleObject)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while creating %s `%s` in namespace `%s`", pretty.APIRule, name, namespace))
		return &gqlschema.APIRule{}, gqlerror.New(err, pretty.APIRules, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	return ar.apiRuleCon.ToGQL(apiRule)
}

func (ar *apiRuleResolver) APIRuleEventSubscription(ctx context.Context, namespace string, serviceName *string) (<-chan gqlschema.ApiRuleEvent, error) {
	channel := make(chan gqlschema.ApiRuleEvent, 1)
	filter := func(apiRule *v1alpha1.APIRule) bool {
		if serviceName == nil {
			return apiRule != nil && apiRule.Namespace == namespace
		}
		return apiRule != nil && apiRule.Namespace == namespace && apiRule.Spec.Service.Name == serviceName
	}

	apiRuleListener := listener.NewApiRule(channel, filter, ar.apiRuleCon, ApiRuleUnstructuredExtractor{})

	ar.apiRuleSvc.Subscribe(apiRuleListener)
	go func() {
		defer close(channel)
		defer ar.apiRuleSvc.Unsubscribe(apiRuleListener)
		<-ctx.Done()
	}()

	return channel, nil
}

func (ar *apiRuleResolver) UpdateAPIRule(ctx context.Context, name string, namespace string, params gqlschema.APIRuleInput) (*gqlschema.APIRule, error) {
	apiRuleObject := ar.apiRuleCon.ToApiRule(name, namespace, params)

	apiRule, err := ar.apiRuleSvc.Update(apiRuleObject)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while editing %s `%s` in namespace `%s`", pretty.APIRule, name, namespace))
		return &gqlschema.APIRule{}, gqlerror.New(err, pretty.APIRules, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	return ar.apiRuleCon.ToGQL(apiRule)
}

func (ar *apiRuleResolver) DeleteAPIRule(ctx context.Context, name string, namespace string) (*gqlschema.APIRule, error) {
	apiRuleObj, err := ar.apiRuleSvc.Find(name, namespace)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting %s `%s` in namespace `%s`", pretty.APIRule, name, namespace))
		return nil, gqlerror.New(err, pretty.APIRules, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	apiRuleCopy := apiRuleObj.DeepCopy()
	err = ar.apiRuleSvc.Delete(name, namespace)

	if err != nil {
		glog.Error(errors.Wrapf(err, "while deleting %s `%s` from namespace `%s`", pretty.APIRule, name, namespace))
		return nil, gqlerror.New(err, pretty.APIRules, gqlerror.WithName(name), gqlerror.WithNamespace(namespace))
	}

	return ar.apiRuleCon.ToGQL(apiRuleCopy)
}
