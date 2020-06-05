package ui

import (
	"fmt"

	"github.com/kyma-project/kyma/common/microfrontend-client/pkg/apis/ui/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/ui/pretty"
	"k8s.io/client-go/tools/cache"
)

type clusterMicroFrontendService struct {
	informer cache.SharedIndexInformer
}

func newClusterMicroFrontendService(informer cache.SharedIndexInformer) *clusterMicroFrontendService {
	return &clusterMicroFrontendService{
		informer: informer,
	}
}

func (svc *clusterMicroFrontendService) List() ([]*v1alpha1.ClusterMicroFrontend, error) {
	items := svc.informer.GetStore().List()

	var clusterMicroFrontends []*v1alpha1.ClusterMicroFrontend
	for _, item := range items {
		clusterMicroFrontend, ok := item.(*v1alpha1.ClusterMicroFrontend)
		if !ok {
			return nil, fmt.Errorf("Incorrect item type: %T, should be: *%s", item, pretty.ClusterMicroFrontend)
		}
		clusterMicroFrontends = append(clusterMicroFrontends, clusterMicroFrontend)
	}

	return clusterMicroFrontends, nil
}
