package mapping

import (
	"fmt"

	"github.com/kyma-project/kyma/components/application-connector/application-broker/pkg/apis/applicationconnector/v1alpha1"
	"k8s.io/client-go/tools/cache"
)

// mappingService provides services for ApplicationMappings
type mappingService struct {
	emInformer cache.SharedIndexInformer
}

func newMappingService(informer cache.SharedIndexInformer) *mappingService {
	return &mappingService{
		emInformer: informer,
	}
}

// ListApplicationMappings lists ApplicationMappings in the given Namespace
func (svc *mappingService) ListApplicationMappings(application string) ([]*v1alpha1.ApplicationMapping, error) {
	items, err := svc.emInformer.GetIndexer().ByIndex(cache.NamespaceIndex, application)
	if err != nil {
		return []*v1alpha1.ApplicationMapping{}, err
	}
	var result []*v1alpha1.ApplicationMapping
	for _, item := range items {
		em, ok := item.(*v1alpha1.ApplicationMapping)
		if !ok {
			return nil, fmt.Errorf("unexpected item type: %T, should be *v1alpha1.ApplicationMapping", item)
		}
		result = append(result, em)
	}

	return result, nil
}
