package ui

import (
	"github.com/kyma-project/kyma/common/microfrontend-client/pkg/apis/ui/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/ui/extractor"
	res "github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
	"k8s.io/client-go/tools/cache"
)

type clusterMicroFrontendService struct {
	informer  cache.SharedIndexInformer
	extractor extractor.CMFUnstructuredExtractor
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
		clusterMicroFrontend, err := res.ToUnstructured(item)
		if err != nil {
			return nil, err
		}
		formattedCMF, err := svc.extractor.FromUnstructured(clusterMicroFrontend)
		if err != nil {
			return nil, err
		}
		clusterMicroFrontends = append(clusterMicroFrontends, formattedCMF)
	}

	return clusterMicroFrontends, nil
}
