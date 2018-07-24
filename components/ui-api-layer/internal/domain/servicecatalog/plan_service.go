package servicecatalog

import (
	"errors"
	"fmt"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"k8s.io/client-go/tools/cache"
)

type planService struct {
	informer cache.SharedIndexInformer
}

func newPlanService(informer cache.SharedIndexInformer) *planService {
	informer.AddIndexers(cache.Indexers{
		"relatedServiceClassName": func(obj interface{}) ([]string, error) {
			servicePlan, ok := obj.(*v1beta1.ClusterServicePlan)
			if !ok {
				return nil, errors.New("Cannot convert item")
			}

			return []string{servicePlan.Spec.ClusterServiceClassRef.Name}, nil
		},
		"classNameAndPlanExternalName": func(obj interface{}) ([]string, error) {
			servicePlan, ok := obj.(*v1beta1.ClusterServicePlan)
			if !ok {
				return nil, errors.New("Cannot convert item")
			}

			str := planIndexKey(servicePlan.Spec.ClusterServiceClassRef.Name, servicePlan.Spec.ExternalName)
			return []string{str}, nil
		},
	})

	return &planService{
		informer: informer,
	}
}

func (svc *planService) Find(name string) (*v1beta1.ClusterServicePlan, error) {
	item, exists, err := svc.informer.GetStore().GetByKey(name)
	if err != nil || !exists {
		return nil, err
	}

	servicePlan, ok := item.(*v1beta1.ClusterServicePlan)
	if !ok {
		return nil, fmt.Errorf("Incorrect item type: %T, should be: *ClusterServicePlan", item)
	}

	return servicePlan, nil
}

func (svc *planService) FindByExternalNameForClass(planExternalName, className string) (*v1beta1.ClusterServicePlan, error) {
	items, err := svc.informer.GetIndexer().ByIndex("classNameAndPlanExternalName", planIndexKey(className, planExternalName))
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
	servicePlan, ok := item.(*v1beta1.ClusterServicePlan)
	if !ok {
		return nil, fmt.Errorf("Incorrect item type: %T, should be: *ClusterServicePlan", item)
	}

	return servicePlan, nil
}

func (svc *planService) ListForClass(name string) ([]*v1beta1.ClusterServicePlan, error) {
	plans, err := svc.informer.GetIndexer().ByIndex("relatedServiceClassName", name)
	if err != nil {
		return nil, err
	}

	var servicePlans []*v1beta1.ClusterServicePlan
	for _, item := range plans {
		servicePlan, ok := item.(*v1beta1.ClusterServicePlan)
		if !ok {
			return nil, fmt.Errorf("Incorrect item type: %T, should be: *ClusterServicePlan", item)
		}

		servicePlans = append(servicePlans, servicePlan)
	}

	return servicePlans, nil
}

func planIndexKey(planExternalName, className string) string {
	return fmt.Sprintf("%s/%s", className, planExternalName)
}
