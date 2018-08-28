package k8s

import (
	"fmt"

	"github.com/pkg/errors"
	api "k8s.io/api/apps/v1beta2"
	"k8s.io/client-go/tools/cache"
)

type deploymentService struct {
	informer cache.SharedIndexInformer
}

func newDeploymentService(informer cache.SharedIndexInformer) *deploymentService {
	svc := &deploymentService{
		informer: informer,
	}

	informer.AddIndexers(cache.Indexers{
		"functionFilter": func(obj interface{}) ([]string, error) {
			deployment, err := svc.toDeployment(obj)
			if err != nil {
				return nil, errors.Wrapf(err, "while indexing by `functionFilter`")
			}

			_, isFunction := deployment.Labels["function"]
			key := fmt.Sprintf("%s/%t", deployment.Namespace, isFunction)
			return []string{key}, nil
		},
	})

	return svc
}

func (svc *deploymentService) Find(name string, environment string) (*api.Deployment, error) {
	key := fmt.Sprintf("%s/%s", environment, name)

	item, exists, err := svc.informer.GetStore().GetByKey(key)
	if err != nil || !exists {
		return nil, err
	}

	deploy, ok := item.(*api.Deployment)
	if !ok {
		return nil, fmt.Errorf("incorrect item type: %T, should be: *v1beta2.Deployment", item)
	}

	return deploy, nil
}

func (svc *deploymentService) List(environment string) ([]*api.Deployment, error) {
	items, err := svc.informer.GetIndexer().ByIndex("namespace", environment)
	if err != nil {
		return nil, err
	}

	return svc.toDeployments(items)
}

func (svc *deploymentService) ListWithoutFunctions(environment string) ([]*api.Deployment, error) {
	key := fmt.Sprintf("%s/false", environment)
	items, err := svc.informer.GetIndexer().ByIndex("functionFilter", key)
	if err != nil {
		return nil, err
	}

	return svc.toDeployments(items)
}

func (svc *deploymentService) toDeployments(items []interface{}) ([]*api.Deployment, error) {
	var deployments []*api.Deployment
	for _, item := range items {
		deployment, err := svc.toDeployment(item)
		if err != nil {
			return nil, err
		}

		deployments = append(deployments, deployment)
	}

	return deployments, nil
}

func (svc *deploymentService) toDeployment(item interface{}) (*api.Deployment, error) {
	deployment, ok := item.(*api.Deployment)
	if !ok {
		return nil, fmt.Errorf("incorrect item type: %T, should be: *Deployment", item)
	}

	return deployment, nil
}
