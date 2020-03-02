package servicecatalog

import (
	"fmt"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/shared"
	"github.com/pkg/errors"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"k8s.io/client-go/tools/cache"
)

type servicePlanService struct {
	informer        cache.SharedIndexInformer
	rafterRetriever shared.RafterRetriever
}

func newServicePlanService(informer cache.SharedIndexInformer, rafterRetriever shared.RafterRetriever) (*servicePlanService, error) {
	err := informer.AddIndexers(cache.Indexers{
		"relatedServiceClassName": func(obj interface{}) ([]string, error) {
			entity, ok := obj.(*v1beta1.ServicePlan)
			if !ok {
				return nil, errors.New("Cannot convert item")
			}

			return []string{fmt.Sprintf("%s/%s", entity.Namespace, entity.Spec.ServiceClassRef.Name)}, nil
		},
		"classNameAndPlanExternalName": func(obj interface{}) ([]string, error) {
			entity, ok := obj.(*v1beta1.ServicePlan)
			if !ok {
				return nil, errors.New("Cannot convert item")
			}

			str := servicePlanIndexKey(entity.Namespace, entity.Spec.ServiceClassRef.Name, entity.Spec.ExternalName)
			return []string{str}, nil
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "while adding indexers")
	}

	return &servicePlanService{
		informer:        informer,
		rafterRetriever: rafterRetriever,
	}, nil
}

func (svc *servicePlanService) Find(name, namespace string) (*v1beta1.ServicePlan, error) {
	key := fmt.Sprintf("%s/%s", namespace, name)
	item, exists, err := svc.informer.GetStore().GetByKey(key)
	if err != nil || !exists {
		return nil, err
	}

	servicePlan, ok := item.(*v1beta1.ServicePlan)
	if !ok {
		return nil, fmt.Errorf("Incorrect item type: %T, should be: *ServicePlan", item)
	}

	return servicePlan, nil
}

func (svc *servicePlanService) FindByExternalName(planExternalName, className, namespace string) (*v1beta1.ServicePlan, error) {
	items, err := svc.informer.GetIndexer().ByIndex("classNameAndPlanExternalName", servicePlanIndexKey(namespace, className, planExternalName))
	if err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return nil, nil
	}

	if len(items) > 1 {
		return nil, fmt.Errorf("Multiple ServicePlan resources with the same externalName %s", planExternalName)
	}

	item := items[0]
	servicePlan, ok := item.(*v1beta1.ServicePlan)
	if !ok {
		return nil, fmt.Errorf("Incorrect item type: %T, should be: *ServicePlan", item)
	}

	return servicePlan, nil
}

func (svc *servicePlanService) ListForServiceClass(name string, namespace string) ([]*v1beta1.ServicePlan, error) {
	key := fmt.Sprintf("%s/%s", namespace, name)
	plans, err := svc.informer.GetIndexer().ByIndex("relatedServiceClassName", key)
	if err != nil {
		return nil, err
	}

	var servicePlans []*v1beta1.ServicePlan
	for _, item := range plans {
		servicePlan, ok := item.(*v1beta1.ServicePlan)
		if !ok {
			return nil, fmt.Errorf("Incorrect item type: %T, should be: *ServicePlan", item)
		}
		// servicePlan.AssetGroup = svc.getAssetGroup(name, namespace)
		servicePlans = append(servicePlans, servicePlan)
	}

	return servicePlans, nil
}

func servicePlanIndexKey(namespace, planExternalName, className string) string {
	return fmt.Sprintf("%s/%s/%s", namespace, className, planExternalName)
}
