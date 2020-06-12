package apigateway

import (
	"github.com/kyma-incubator/api-gateway/api/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
	"k8s.io/apimachinery/pkg/runtime/schema"
)


type APIRuleList []*v1alpha1.APIRule

func (l *APIRuleList) Append() interface{} {
	e := &v1alpha1.APIRule{}
	*l = append(*l, e)
	return e
}

var apiRulesKind = "APIRule"
var apiRulesGroupVersionResource = schema.GroupVersionResource{
	Version:  v1alpha1.GroupVersion.Version,
	Group:    v1alpha1.GroupVersion.Group,
	Resource: "apirules",
}

type Service struct {
	*resource.Service
}

func NewService(serviceFactory *resource.ServiceFactory) *Service {
	return &Service{
		Service:  serviceFactory.ForResource(apiRulesGroupVersionResource),
	}
}

func (svc *Service) Find(name, namespace string) (*v1alpha1.APIRule, error) {
	var result *v1alpha1.APIRule
	err := svc.GetInNamespace(name, namespace, &result)
	return result, err
}

func (svc *Service) Delete(name, namespace string) (*v1alpha1.APIRule, error) {
	var result *v1alpha1.APIRule
	err := svc.Service.DeleteInNamespace(namespace, name, result)
	return result, err
}

func (svc *Service) Create(apiRule *v1alpha1.APIRule) (*v1alpha1.APIRule, error) {
	var result *v1alpha1.APIRule
	err := svc.Service.Create(apiRule, result)
	return result, err
}

func (svc *Service) Subscribe(handler resource.EventHandlerProvider) resource.Unsubscribe {
	return svc.Subscribe(handler)
}

func (svc *Service) Update(name, namespace string, newSpec v1alpha1.APIRuleSpec) (*v1alpha1.APIRule, error) {
	var result *v1alpha1.APIRule
	err := svc.Service.Update(name, namespace, result, func() error {
		result.Spec = newSpec
		return nil
	})
	return result, err
}

func (svc *Service) List(namespace string, serviceName *string, hostname *string) ([]*v1alpha1.APIRule, error) {
	items := APIRuleList{}
	err := svc.ListInNamespace(namespace, &items)
	if err != nil {
		return nil, err
	}

	var apiRules []*v1alpha1.APIRule
	for _, item := range items {
		match := true
		if serviceName != nil {
			if *serviceName != *item.Spec.Service.Name {
				match = false
			}
		}
		if hostname != nil {
			if *hostname != *item.Spec.Service.Host {
				match = false
			}
		}

		if match {
			apiRules = append(apiRules, item)
		}
	}

	return apiRules, nil
}

func NewEventHandler(channel chan<- *gqlschema.APIRuleEvent, filter func(v1alpha1.APIRule) bool) resource.EventHandlerProvider {
	return func() resource.EventHandler {
		return &EventHandler{
			channel: channel,
			filter:  filter,
		}
	}
}

type EventHandler struct {
	channel chan<- *gqlschema.APIRuleEvent
	filter  func(v1alpha1.APIRule) bool
	res     *v1alpha1.APIRule
}

func (h *EventHandler) K8sResource() interface{} {
	return h.res
}

func (h *EventHandler) ShouldNotify() bool {
	return h.filter(*h.res)
}

func (h *EventHandler) Notify(eventType resource.EventType) {
	h.channel <- &gqlschema.APIRuleEvent{
		Type:    eventType,
		APIRule: h.res,
	}
}

