package apigateway

import (
	"github.com/kyma-incubator/api-gateway/api/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
	notifierRes "github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"
	"github.com/pkg/errors"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var apiRulesGroupVersionResource = schema.GroupVersionResource{
	Version:  v1alpha1.GroupVersion.Version,
	Group:    v1alpha1.GroupVersion.Group,
	Resource: "apirules",
}

type Service struct {
	*resource.Service
	notifier notifierRes.Notifier
}

func NewService(serviceFactory *resource.ServiceFactory) *Service {
	notifier := notifierRes.NewNotifier()
	return &Service{
		Service:  serviceFactory.ForResource(apiRulesGroupVersionResource),
		notifier: notifier,
	}
}

func (svc *Service) List(namespace string, serviceName *string, hostname *string) ([]*v1alpha1.APIRule, error) {
	items := make([]*v1alpha1.APIRule, 0)
	err := svc.ListInIndex("namespace", namespace, &items)
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

func (svc *Service) Find(name, namespace string) (*v1alpha1.APIRule, error) {
	var result *v1alpha1.APIRule
	err := svc.FindInNamespace(name, namespace, &result)
	return result, err
}

func (svc *Service) Delete(name, namespace string) error {
	return svc.Client.Namespace(namespace).Delete(name, &metav1.DeleteOptions{})
}

var apiRuleTypeMeta = metav1.TypeMeta{
	Kind:       "APIRule",
	APIVersion: "gateway.kyma-project.io/v1alpha1",
}

func (svc *Service) Create(apiRule *v1alpha1.APIRule) (*v1alpha1.APIRule, error) {
	apiRule.TypeMeta = apiRuleTypeMeta

	u, err := toUnstructured(apiRule)
	if err != nil {
		return &v1alpha1.APIRule{}, err
	}

	created, err := svc.Client.Namespace(apiRule.ObjectMeta.Namespace).Create(u, metav1.CreateOptions{})
	if err != nil {
		return &v1alpha1.APIRule{}, err
	}
	return fromUnstructured(created)
}

func (svc *Service) Subscribe(listener notifierRes.Listener) {
	svc.notifier.AddListener(listener)
}

func (svc *Service) Unsubscribe(listener notifierRes.Listener) {
	svc.notifier.DeleteListener(listener)
}

func (svc *Service) Update(apiRule *v1alpha1.APIRule) (*v1alpha1.APIRule, error) {
	oldApiRule, err := svc.Find(apiRule.ObjectMeta.Name, apiRule.ObjectMeta.Namespace)
	if err != nil {
		return nil, errors.Wrapf(err, "while finding APIRule %s", apiRule.ObjectMeta.Name)
	}

	if oldApiRule == nil {
		return nil, apiErrors.NewNotFound(schema.GroupResource{
			Group:    apiRulesGroupVersionResource.Group,
			Resource: apiRulesGroupVersionResource.Resource,
		}, apiRule.ObjectMeta.Name)
	}
	apiRule.ObjectMeta.ResourceVersion = oldApiRule.ObjectMeta.ResourceVersion
	apiRule.TypeMeta = apiRuleTypeMeta

	u, err := toUnstructured(apiRule)
	if err != nil {
		return &v1alpha1.APIRule{}, err
	}

	updated, err := svc.Client.Namespace(apiRule.ObjectMeta.Namespace).Update(u, metav1.UpdateOptions{})
	if err != nil {
		return &v1alpha1.APIRule{}, err
	}
	return fromUnstructured(updated)
}
