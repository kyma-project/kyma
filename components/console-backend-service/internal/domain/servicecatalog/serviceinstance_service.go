package servicecatalog

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalog/status"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/pager"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
)

type serviceInstanceService struct {
	informer    cache.SharedIndexInformer
	client      clientset.Interface
	notifier    notifier
	instanceExt status.InstanceExtractor
}

func newServiceInstanceService(informer cache.SharedIndexInformer, client clientset.Interface) (*serviceInstanceService, error) {
	instanceExt := status.InstanceExtractor{}
	err := informer.AddIndexers(cache.Indexers{
		"externalClusterServiceClassName": func(obj interface{}) ([]string, error) {
			serviceInstance, ok := obj.(*v1beta1.ServiceInstance)
			if !ok {
				return nil, fmt.Errorf("Cannot convert item")
			}

			return []string{serviceInstance.Spec.PlanReference.ClusterServiceClassExternalName}, nil
		},
		"clusterServiceClassName": func(obj interface{}) ([]string, error) {
			serviceInstance, ok := obj.(*v1beta1.ServiceInstance)
			if !ok {
				return nil, fmt.Errorf("Cannot convert item")
			}
			return []string{serviceInstance.Spec.PlanReference.ClusterServiceClassName}, nil
		},
		"ns/externalClusterServiceClassName": func(obj interface{}) ([]string, error) {
			serviceInstance, ok := obj.(*v1beta1.ServiceInstance)
			if !ok {
				return nil, fmt.Errorf("Cannot convert item")
			}

			return []string{fmt.Sprintf("%s/%s", serviceInstance.Namespace, serviceInstance.Spec.PlanReference.ClusterServiceClassExternalName)}, nil
		},
		"ns/clusterServiceClassName": func(obj interface{}) ([]string, error) {
			serviceInstance, ok := obj.(*v1beta1.ServiceInstance)
			if !ok {
				return nil, fmt.Errorf("Cannot convert item")
			}
			return []string{fmt.Sprintf("%s/%s", serviceInstance.Namespace, serviceInstance.Spec.PlanReference.ClusterServiceClassName)}, nil
		},
		"externalServiceClassName": func(obj interface{}) ([]string, error) {
			serviceInstance, ok := obj.(*v1beta1.ServiceInstance)
			if !ok {
				return nil, fmt.Errorf("Cannot convert item")
			}

			return []string{fmt.Sprintf("%s/%s", serviceInstance.Namespace, serviceInstance.Spec.PlanReference.ServiceClassExternalName)}, nil
		},
		"serviceClassName": func(obj interface{}) ([]string, error) {
			serviceInstance, ok := obj.(*v1beta1.ServiceInstance)
			if !ok {
				return nil, fmt.Errorf("Cannot convert item")
			}
			return []string{fmt.Sprintf("%s/%s", serviceInstance.Namespace, serviceInstance.Spec.PlanReference.ServiceClassName)}, nil
		},
		"statusType": func(obj interface{}) ([]string, error) {
			serviceInstance, ok := obj.(*v1beta1.ServiceInstance)
			if !ok {
				return nil, fmt.Errorf("Cannot convert item")
			}

			if obj == nil {
				return nil, fmt.Errorf("Nil reference")
			}

			key := fmt.Sprintf("%s/%s", serviceInstance.ObjectMeta.Namespace, instanceExt.Status(*serviceInstance).Type)
			return []string{key}, nil
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "while adding indexers")
	}

	notifier := resource.NewNotifier()
	informer.AddEventHandler(notifier)

	return &serviceInstanceService{
		informer:    informer,
		client:      client,
		notifier:    notifier,
		instanceExt: instanceExt,
	}, nil
}

func (svc *serviceInstanceService) Find(name, namespace string) (*v1beta1.ServiceInstance, error) {
	key := fmt.Sprintf("%s/%s", namespace, name)

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

func (svc *serviceInstanceService) List(namespace string, pagingParams pager.PagingParams) ([]*v1beta1.ServiceInstance, error) {
	items, err := pager.FromIndexer(svc.informer.GetIndexer(), "namespace", namespace).Limit(pagingParams)
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

func (svc *serviceInstanceService) ListForStatus(namespace string, pagingParams pager.PagingParams, status *status.ServiceInstanceStatusType) ([]*v1beta1.ServiceInstance, error) {
	key := fmt.Sprintf("%s/%s", namespace, *status)
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

func (svc *serviceInstanceService) ListForClusterServiceClass(className, externalClassName string, namespace *string) ([]*v1beta1.ServiceInstance, error) {
	indexName := func(indexName string) string {
		if namespace == nil {
			return indexName
		}

		return fmt.Sprintf("ns/%s", indexName)
	}

	indexValue := func(value string) string {
		if namespace == nil {
			return value
		}

		return fmt.Sprintf("%s/%s", *namespace, value)
	}

	indexer := svc.informer.GetIndexer()
	itemsByClassName, err := indexer.ByIndex(indexName("clusterServiceClassName"), indexValue(className))
	if err != nil {
		return nil, err
	}

	itemsByExternalClassName, err := indexer.ByIndex(indexName("externalClusterServiceClassName"), indexValue(externalClassName))
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

func (svc *serviceInstanceService) ListForServiceClass(className, externalClassName string, namespace string) ([]*v1beta1.ServiceInstance, error) {
	indexer := svc.informer.GetIndexer()
	itemsByClassName, err := indexer.ByIndex("serviceClassName", fmt.Sprintf("%s/%s", namespace, className))
	if err != nil {
		return nil, err
	}

	itemsByExternalClassName, err := indexer.ByIndex("externalServiceClassName", fmt.Sprintf("%s/%s", namespace, externalClassName))
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

type instanceCreateResourceRef struct {
	ExternalName string
	ClusterWide  bool
}

type serviceInstanceCreateParameters struct {
	Name      string
	Namespace string
	Labels    []string
	PlanRef   instanceCreateResourceRef
	ClassRef  instanceCreateResourceRef
	Schema    map[string]interface{}
}

func (svc *serviceInstanceService) Create(params serviceInstanceCreateParameters) (*v1beta1.ServiceInstance, error) {
	specParameters, err := svc.createInstanceParameters(params.Schema)
	if err != nil {
		return nil, errors.Wrap(err, "while creating spec parameters")
	}

	filtered := svc.filterTags(params.Labels)
	annotations := map[string]string{
		"tags": strings.Join(filtered, ","),
	}

	var clusterServicePlanExternalName string
	var servicePlanExternalName string

	if params.PlanRef.ClusterWide {
		clusterServicePlanExternalName = params.PlanRef.ExternalName
	} else {
		servicePlanExternalName = params.PlanRef.ExternalName
	}

	var clusterServiceClassExternalName string
	var serviceClassExternalName string

	if params.ClassRef.ClusterWide {
		clusterServiceClassExternalName = params.ClassRef.ExternalName
	} else {
		serviceClassExternalName = params.ClassRef.ExternalName
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
				ClusterServiceClassExternalName: clusterServiceClassExternalName,
				ClusterServicePlanExternalName:  clusterServicePlanExternalName,
				ServiceClassExternalName:        serviceClassExternalName,
				ServicePlanExternalName:         servicePlanExternalName,
			},
			Parameters: specParameters,
		},
	}

	return svc.client.ServicecatalogV1beta1().ServiceInstances(params.Namespace).Create(context.Background(), &instance, v1.CreateOptions{})
}

func (svc *serviceInstanceService) Delete(name, namespace string) error {
	return svc.client.ServicecatalogV1beta1().ServiceInstances(namespace).Delete(context.Background(), name, v1.DeleteOptions{})
}

func (svc *serviceInstanceService) IsBindableWithClusterRefs(relatedClass *v1beta1.ClusterServiceClass, relatedPlan *v1beta1.ClusterServicePlan) bool {
	if relatedPlan != nil && relatedPlan.Spec.Bindable != nil {
		return *relatedPlan.Spec.Bindable
	}

	if relatedClass != nil {
		return relatedClass.Spec.Bindable
	}

	return false
}

func (svc *serviceInstanceService) IsBindableWithLocalRefs(relatedClass *v1beta1.ServiceClass, relatedPlan *v1beta1.ServicePlan) bool {
	if relatedPlan != nil && relatedPlan.Spec.Bindable != nil {
		return *relatedPlan.Spec.Bindable
	}

	if relatedClass != nil {
		return relatedClass.Spec.Bindable
	}

	return false
}

func (svc *serviceInstanceService) Subscribe(listener resource.Listener) {
	svc.notifier.AddListener(listener)
}

func (svc *serviceInstanceService) Unsubscribe(listener resource.Listener) {
	svc.notifier.DeleteListener(listener)
}

func (svc *serviceInstanceService) createInstanceParameters(schema map[string]interface{}) (*runtime.RawExtension, error) {
	parameters := runtime.DeepCopyJSON(schema)

	byteArray, err := json.Marshal(parameters)
	if err != nil {
		return nil, errors.Wrap(err, "while marshalling parameters")
	}

	return &runtime.RawExtension{
		Raw: byteArray,
	}, nil
}

func (svc *serviceInstanceService) uniqueInstances(items []*v1beta1.ServiceInstance) []*v1beta1.ServiceInstance {
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

func (svc *serviceInstanceService) filterTags(labels []string) []string {
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
