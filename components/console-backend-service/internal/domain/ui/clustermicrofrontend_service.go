package ui

import (
	"fmt"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/ui/extractor"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/kyma-project/kyma/common/microfrontend-client/pkg/apis/ui/v1alpha1"
	"k8s.io/client-go/tools/cache"
)

type clusterMicroFrontendService struct {
	informer cache.SharedIndexInformer
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
		clusterMicroFrontend, ok := item.(*unstructured.Unstructured)
		if !ok {
			return nil, fmt.Errorf("incorrect item type: %T, should be: *unstructured.Unstructured", item)
		}
		formattedCMF, err := svc.extractor.FromUnstructured(clusterMicroFrontend)
		if err != nil {
			return nil, err
		}
		clusterMicroFrontends = append(clusterMicroFrontends, formattedCMF)
	}

	return clusterMicroFrontends, nil
}
