package servicecatalog

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/servicecatalog/status"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/pager"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/jsoncopy"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/resource"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
)

type instanceService struct {
	informer    cache.SharedIndexInformer
	client      clientset.Interface
	notifier    notifier
	instanceExt status.InstanceExtractor
}

func newInstanceService(informer cache.SharedIndexInformer, client clientset.Interface) *instanceService {
	instanceExt := status.InstanceExtractor{}
	informer.AddIndexers(cache.Indexers{
		"relatedExternalClassName": func(obj interface{}) ([]string, error) {
			serviceInstance, ok := obj.(*v1beta1.ServiceInstance)
			if !ok {
				return nil, fmt.Errorf("Cannot convert item")
			}

			return []string{serviceInstance.Spec.PlanReference.ClusterServiceClassExternalName}, nil
		},
		"relatedClassName": func(obj interface{}) ([]string, error) {
			serviceInstance, ok := obj.(*v1beta1.ServiceInstance)
			if !ok {
				return nil, fmt.Errorf("Cannot convert item")
			}
			return []string{serviceInstance.Spec.PlanReference.ClusterServiceClassName}, nil
		},
		"statusType": func(obj interface{}) ([]string, error) {
			serviceInstance, ok := obj.(*v1beta1.ServiceInstance)
			if !ok {
				return nil, fmt.Errorf("Cannot convert item")
			}

			key := fmt.Sprintf("%s/%s", serviceInstance.ObjectMeta.Namespace, instanceExt.Status(serviceInstance).Type)
			return []string{key}, nil
		},
	})

	notifier := resource.NewNotifier()
	informer.AddEventHandler(notifier)

	return &instanceService{
		informer:    informer,
		client:      client,
		notifier:    notifier,
		instanceExt: instanceExt,
	}
}

func (svc *instanceService) Find(name, environment string) (*v1beta1.ServiceInstance, error) {
	key := fmt.Sprintf("%s/%s", environment, name)

	item, exists, err := svc.informer.GetStore().GetByKey(key)
	if err != nil || !exists {
		return nil, err
	}

	serviceInstance, ok := item.(*v1beta1.ServiceInstance)
	if !ok {
		return nil, fmt.Errorf("Incorrect item type: %T, should be: *ServiceInstance", item)
	}

	return serviceInstance, nil
}

func (svc *instanceService) List(environment string, pagingParams pager.PagingParams) ([]*v1beta1.ServiceInstance, error) {
	items, err := pager.FromIndexer(svc.informer.GetIndexer(), "namespace", environment).Limit(pagingParams)
	if err != nil {
		return nil, err
	}

	var serviceInstances []*v1beta1.ServiceInstance
	for _, item := range items {
		serviceInstance, ok := item.(*v1beta1.ServiceInstance)
		if !ok {
			return nil, fmt.Errorf("Incorrect item type: %T, should be: *ServiceInstance", item)
		}
		serviceInstances = append(serviceInstances, serviceInstance)
	}

	return serviceInstances, nil
}

func (svc *instanceService) ListForStatus(environment string, pagingParams pager.PagingParams, status *status.ServiceInstanceStatusType) ([]*v1beta1.ServiceInstance, error) {
	key := fmt.Sprintf("%s/%s", environment, *status)
	items, err := pager.FromIndexer(svc.informer.GetIndexer(), "statusType", key).Limit(pagingParams)
	if err != nil {
		return nil, err
	}

	var serviceInstances []*v1beta1.ServiceInstance
	for _, item := range items {
		serviceInstance, ok := item.(*v1beta1.ServiceInstance)
		if !ok {
			return nil, fmt.Errorf("Incorrect item type: %T, should be: *ServiceInstance", item)
		}
		serviceInstances = append(serviceInstances, serviceInstance)
	}

	return serviceInstances, nil
}

func (svc *instanceService) ListForClass(className, externalClassName string) ([]*v1beta1.ServiceInstance, error) {
	indexer := svc.informer.GetIndexer()
	itemsByClassName, err := indexer.ByIndex("relatedClassName", className)
	if err != nil {
		return nil, err
	}

	itemsByExternalClassName, err := indexer.ByIndex("relatedExternalClassName", externalClassName)
	if err != nil {
		return nil, err
	}

	items := append(itemsByClassName, itemsByExternalClassName...)
	var serviceInstances []*v1beta1.ServiceInstance
	for _, item := range items {
		serviceInstance, ok := item.(*v1beta1.ServiceInstance)
		if !ok {
			return nil, fmt.Errorf("Incorrect item type: %T, should be: *ServiceInstance", item)
		}

		serviceInstances = append(serviceInstances, serviceInstance)
	}

	return svc.uniqueInstances(serviceInstances), nil
}

type instanceCreateParameters struct {
	Name                     string
	Namespace                string
	Labels                   []string
	ExternalServicePlanName  string
	ExternalServiceClassName string
	Schema                   map[string]interface{}
}

func (svc *instanceService) Create(params instanceCreateParameters) (*v1beta1.ServiceInstance, error) {
	specParameters, err := svc.createInstanceParameters(params.Schema)
	if err != nil {
		return nil, errors.Wrap(err, "while creating spec parameters")
	}

	filtered := svc.filterTags(params.Labels)
	annotations := map[string]string{
		"tags": strings.Join(filtered, ","),
	}

	instance := v1beta1.ServiceInstance{
		TypeMeta: v1.TypeMeta{
			APIVersion: "servicecatalog.k8s.io/v1beta1",
			Kind:       "ServiceInstance",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:        params.Name,
			Namespace:   params.Namespace,
			Annotations: annotations,
		},
		Spec: v1beta1.ServiceInstanceSpec{
			PlanReference: v1beta1.PlanReference{
				ClusterServiceClassExternalName: params.ExternalServiceClassName,
				ClusterServicePlanExternalName:  params.ExternalServicePlanName,
			},
			Parameters: specParameters,
		},
	}

	return svc.client.ServicecatalogV1beta1().ServiceInstances(params.Namespace).Create(&instance)
}

func (svc *instanceService) Delete(name, namespace string) error {
	return svc.client.ServicecatalogV1beta1().ServiceInstances(namespace).Delete(name, nil)
}

func (svc *instanceService) IsBindable(relatedClass *v1beta1.ClusterServiceClass, relatedPlan *v1beta1.ClusterServicePlan) bool {
	if relatedPlan != nil && relatedPlan.Spec.Bindable != nil {
		return *relatedPlan.Spec.Bindable
	}

	if relatedClass != nil {
		return relatedClass.Spec.Bindable
	}

	return false
}

func (svc *instanceService) Subscribe(listener resource.Listener) {
	svc.notifier.AddListener(listener)
}

func (svc *instanceService) Unsubscribe(listener resource.Listener) {
	svc.notifier.DeleteListener(listener)
}

func (svc *instanceService) createInstanceParameters(schema map[string]interface{}) (*runtime.RawExtension, error) {
	parameters := jsoncopy.DeepCopyJSON(schema)

	byteArray, err := json.Marshal(parameters)
	if err != nil {
		return nil, errors.Wrap(err, "while marshalling parameters")
	}

	return &runtime.RawExtension{
		Raw: byteArray,
	}, nil
}

func (svc *instanceService) uniqueInstances(items []*v1beta1.ServiceInstance) []*v1beta1.ServiceInstance {
	keys := make(map[string]bool)
	var uniqueItems []*v1beta1.ServiceInstance

	for _, item := range items {
		if _, value := keys[item.Name]; !value {
			keys[item.Name] = true
			uniqueItems = append(uniqueItems, item)
		}
	}

	return uniqueItems
}

func (svc *instanceService) filterTags(labels []string) []string {
	r := regexp.MustCompile("[^a-zA-Z0-9 _-]")

	var filtered []string
	for _, v := range labels {
		clean := strings.TrimSpace(r.ReplaceAllString(v, ""))

		if len(clean) > 0 {
			filtered = append(filtered, clean)
		}
	}

	return filtered
}
