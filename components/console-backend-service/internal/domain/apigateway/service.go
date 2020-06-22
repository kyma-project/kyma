package apigateway

import (
	"fmt"

	"github.com/kyma-incubator/api-gateway/api/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/apigateway/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
	notifierRes "github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"
	"github.com/pkg/errors"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"
)

const (
	Namespace_Index                       = "namespace"
	Namespace_Service_Name_Index          = "namespace/serviceName"
	Namespace_Hostname_Index              = "namespace/hostname"
	Namespace_Service_Name_Hostname_Index = "namespace/serviceName/hostname"
)

var apiRulesGroupVersionResource = schema.GroupVersionResource{
	Version:  v1alpha1.GroupVersion.Version,
	Group:    v1alpha1.GroupVersion.Group,
	Resource: "apirules",
}

type Service struct {
	*resource.Service
	notifier  notifierRes.Notifier
	extractor *ApiRuleUnstructuredExtractor
}

func NewService(serviceFactory *resource.ServiceFactory) (*Service, error) {
	svc := &Service{
		Service:   serviceFactory.ForResource(apiRulesGroupVersionResource),
		extractor: &ApiRuleUnstructuredExtractor{},
	}

	err := svc.AddIndexers(cache.Indexers{
		Namespace_Service_Name_Hostname_Index: func(obj interface{}) ([]string, error) {
			entity, err := svc.extractor.Do(obj)
			if err != nil || entity == nil || entity.Spec.Service == nil || entity.Spec.Service.Name == nil || entity.Spec.Service.Host == nil {
				return nil, errors.New("Cannot convert item")
			}
			return []string{fmt.Sprintf("%s/%s/%s", entity.Namespace, *entity.Spec.Service.Name, *entity.Spec.Service.Host)}, nil
		},
		Namespace_Service_Name_Index: func(obj interface{}) ([]string, error) {
			entity, err := svc.extractor.Do(obj)
			if err != nil || entity == nil || entity.Spec.Service == nil || entity.Spec.Service.Name == nil {
				return nil, errors.New("Cannot convert item")
			}
			return []string{fmt.Sprintf("%s/%s", entity.Namespace, *entity.Spec.Service.Name)}, nil
		},
		Namespace_Hostname_Index: func(obj interface{}) ([]string, error) {
			entity, err := svc.extractor.Do(obj)
			if err != nil || entity == nil || entity.Spec.Service == nil || entity.Spec.Service.Host == nil {
				return nil, errors.New("Cannot convert item")
			}
			return []string{fmt.Sprintf("%s/%s", entity.Namespace, *entity.Spec.Service.Host)}, nil
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "while adding indexers")
	}

	notifier := notifierRes.NewNotifier()
	svc.Informer.AddEventHandler(notifier)
	svc.notifier = notifier

	return svc, nil
}

func (svc *Service) List(namespace string, serviceName *string, hostname *string) ([]*v1alpha1.APIRule, error) {
	var items []interface{}
	var err error
	if serviceName != nil && hostname != nil {
		items, err = svc.Informer.GetIndexer().ByIndex(Namespace_Service_Name_Hostname_Index, fmt.Sprintf("%s/%s/%s", namespace, *serviceName, *hostname))
	} else if serviceName != nil {
		items, err = svc.Informer.GetIndexer().ByIndex(Namespace_Service_Name_Index, fmt.Sprintf("%s/%s", namespace, *serviceName))
	} else if hostname != nil {
		items, err = svc.Informer.GetIndexer().ByIndex(Namespace_Hostname_Index, fmt.Sprintf("%s/%s", namespace, *hostname))
	} else {
		items, err = svc.Informer.GetIndexer().ByIndex(Namespace_Index, namespace)
	}
	if err != nil {
		return nil, err
	}

	var apiRules []*v1alpha1.APIRule
	for _, item := range items {
		apiRule, err := svc.extractor.Do(item)
		if err != nil {
			return nil, errors.Wrapf(err, "Incorrect item type: %T, should be: *%s", item, pretty.APIRule)
		}
		apiRules = append(apiRules, apiRule)
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
