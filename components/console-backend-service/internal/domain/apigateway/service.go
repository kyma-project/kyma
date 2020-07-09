package apigateway

import (
	"github.com/kyma-incubator/api-gateway/api/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"
)

var apiRulesKind = "APIRule"
var apiRulesGroupVersionResource = schema.GroupVersionResource{
	Version:  v1alpha1.GroupVersion.Version,
	Group:    v1alpha1.GroupVersion.Group,
	Resource: "apirules",
}

var apiRulesServiceIndex = "service"

func apiRulesServiceIndexKey(namespace string, serviceName *string) string {
	return namespace + ":" + *serviceName
}

var apiRulesHostnameIndex = "hostname"

func apiRulesHostnameIndexKey(namespace string, hostname *string) string {
	return namespace + ":" + *hostname
}

var apiRulesServiceAndHostnameIndex = "service-hostname"

func apiRulesServiceAndHostnameIndexKey(namespace string, serviceName *string, hostname *string) string {
	return namespace + ":" + *serviceName + ":" + *hostname
}

type Service struct {
	*resource.Service
}

func NewService(serviceFactory *resource.GenericServiceFactory) (*resource.GenericService, error) {
	service := serviceFactory.ForResource(apiRulesGroupVersionResource)
	err := service.AddIndexers(cache.Indexers{
		apiRulesServiceIndex: func(obj interface{}) ([]string, error) {
			rule := &v1alpha1.APIRule{}
			err := resource.FromUnstructured(obj.(*unstructured.Unstructured), rule)
			if err != nil {
				return nil, err
			}
			return []string{apiRulesServiceIndexKey(rule.Namespace, rule.Spec.Service.Name)}, nil
		},
		apiRulesHostnameIndex: func(obj interface{}) ([]string, error) {
			rule := &v1alpha1.APIRule{}
			err := resource.FromUnstructured(obj.(*unstructured.Unstructured), rule)
			if err != nil {
				return nil, err
			}
			return []string{apiRulesHostnameIndexKey(rule.Namespace, rule.Spec.Service.Host)}, nil
		},
		apiRulesServiceAndHostnameIndex: func(obj interface{}) ([]string, error) {
			rule := &v1alpha1.APIRule{}
			err := resource.FromUnstructured(obj.(*unstructured.Unstructured), rule)
			if err != nil {
				return nil, err
			}
			return []string{apiRulesServiceAndHostnameIndexKey(rule.Namespace, rule.Spec.Service.Name, rule.Spec.Service.Host)}, nil
		},
	})
	return service, err
}

func NewEventHandler(channel chan<- *gqlschema.APIRuleEvent, filter func(v1alpha1.APIRule) bool) resource.EventHandlerProvider {
	return func() resource.EventHandler {
		return &EventHandler{
			channel: channel,
			filter:  filter,
			res:     &v1alpha1.APIRule{},
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

func (h *EventHandler) Notify(eventType gqlschema.SubscriptionEventType) {
	h.channel <- &gqlschema.APIRuleEvent{
		Type:    eventType,
		APIRule: h.res,
	}
}
