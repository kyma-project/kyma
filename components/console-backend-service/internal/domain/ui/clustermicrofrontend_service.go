package ui

import (
	"fmt"

	uiV1alpha1v "github.com/kyma-project/kyma/common/microfrontend-client/pkg/apis/ui/v1alpha1"
	"k8s.io/client-go/tools/cache"
)

type clusterMicrofrontendService struct {
	informer cache.SharedIndexInformer
}

func newClusterMicrofrontendService(informer cache.SharedIndexInformer) *clusterMicrofrontendService {
	return &clusterMicrofrontendService{
		informer: informer,
	}
}

func (svc *clusterMicrofrontendService) List() ([]*uiV1alpha1v.ClusterMicroFrontend, error) {
	items := svc.informer.GetStore().List()

	var clusterMicrofrontends []*uiV1alpha1v.ClusterMicroFrontend
	for _, item := range items {
		clusterMicrofrontend, ok := item.(*uiV1alpha1v.ClusterMicroFrontend)
		if !ok {
			return nil, fmt.Errorf("Incorrect item type: %T, should be: *ClusterMicrofrontend", item)
		}
		clusterMicrofrontends = append(clusterMicrofrontends, clusterMicrofrontend)
	}

	return clusterMicrofrontends, nil
}
