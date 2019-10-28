package apigateway

import (
	"fmt"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/apigateway/extractor"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"

	"github.com/kyma-incubator/api-gateway/api/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"
	"github.com/pkg/errors"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"
)

type apiRuleService struct {
	informer  cache.SharedIndexInformer
	client    dynamic.NamespaceableResourceInterface
	notifier  resource.Notifier
	extractor extractor.ApiRuleUnstructuredExtractor
}

var apiRuleTypeMeta = metav1.TypeMeta{
	Kind:       "APIRule",
	APIVersion: "gateway.kyma-project.io/v1alpha1",
}

func newApiRuleService(informer cache.SharedIndexInformer, client dynamic.NamespaceableResourceInterface) *apiRuleService {
	notifier := resource.NewNotifier()
	informer.AddEventHandler(notifier)

	return &apiRuleService{
		informer:  informer,
		client:    client,
		notifier:  notifier,
		extractor: extractor.ApiRuleUnstructuredExtractor{},
	}
}

func (svc *apiRuleService) List(namespace string, serviceName *string, hostname *string) ([]*v1alpha1.APIRule, error) {
	items, err := svc.informer.GetIndexer().ByIndex("namespace", namespace)
	if err != nil {
		return nil, err
	}

	var apiRules []*v1alpha1.APIRule
	for _, item := range items {

		apiRule, ok := item.(*unstructured.Unstructured)
		if !ok {
			return nil, fmt.Errorf("incorrect item type: %T, should be: *unstructured.Unstructured", item)
		}

		formattedApiRule, err := svc.extractor.FromUnstructured(apiRule)
		if err != nil {
			return nil, err
		}
		match := true
		if serviceName != nil {
			if serviceName != formattedApiRule.Spec.Service.Name {
				match = false
			}
		}
		if hostname != nil {
			if hostname != formattedApiRule.Spec.Service.Host {
				match = false
			}
		}

		if match {
			apiRules = append(apiRules, formattedApiRule)
		}
	}

	return apiRules, nil
}

func (svc *apiRuleService) Find(name string, namespace string) (*v1alpha1.APIRule, error) {
	key := fmt.Sprintf("%s/%s", namespace, name)
	item, exists, err := svc.informer.GetStore().GetByKey(key)
	if err != nil {
		return nil, errors.Wrapf(err, "while finding APIRule %s", name)
	}

	if !exists {
		return nil, nil
	}

	res, err := svc.extractor.Do(item)
	if err != nil {
		return &v1alpha1.APIRule{}, err
	}

	return res, nil
}

func (svc *apiRuleService) Create(apiRule *v1alpha1.APIRule) (*v1alpha1.APIRule, error) {
	apiRule.TypeMeta = apiRuleTypeMeta

	u, err := svc.extractor.ToUnstructured(apiRule)
	if err != nil {
		return &v1alpha1.APIRule{}, err
	}

	created, err := svc.client.Namespace(apiRule.ObjectMeta.Namespace).Create(u, metav1.CreateOptions{})
	if err != nil {
		return &v1alpha1.APIRule{}, err
	}
	return svc.extractor.FromUnstructured(created)
}

func (svc *apiRuleService) Subscribe(listener resource.Listener) {
	svc.notifier.AddListener(listener)
}

func (svc *apiRuleService) Unsubscribe(listener resource.Listener) {
	svc.notifier.DeleteListener(listener)
}

func (svc *apiRuleService) Update(apiRule *v1alpha1.APIRule) (*v1alpha1.APIRule, error) {
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
	apiRule.ObjectMeta.ResourceVersion = apiRule.ObjectMeta.ResourceVersion
	apiRule.TypeMeta = apiRuleTypeMeta

	u, err := svc.extractor.ToUnstructured(apiRule)
	if err != nil {
		return &v1alpha1.APIRule{}, err
	}

	updated, err := svc.client.Namespace(apiRule.ObjectMeta.Namespace).Update(u, metav1.UpdateOptions{})
	if err != nil {
		return &v1alpha1.APIRule{}, err
	}
	return svc.extractor.FromUnstructured(updated)
}

func (svc *apiRuleService) Delete(name string, namespace string) error {
	return svc.client.Namespace(namespace).Delete(name, nil)
}
