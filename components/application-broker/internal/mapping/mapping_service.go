package mapping

import (
	"fmt"

	"github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	"k8s.io/client-go/tools/cache"
)

// mappingService provides services for EnvironmentMappings
type mappingService struct {
	emInformer cache.SharedIndexInformer
}

func newMappingService(informer cache.SharedIndexInformer) *mappingService {
	return &mappingService{
		emInformer: informer,
	}
}

// ListEnvironmentMappings lists EnvironmentMappings in the given Namespace
func (svc *mappingService) ListEnvironmentMappings(environment string) ([]*v1alpha1.EnvironmentMapping, error) {
	items, err := svc.emInformer.GetIndexer().ByIndex(cache.NamespaceIndex, environment)
	if err != nil {
		return []*v1alpha1.EnvironmentMapping{}, err
	}
	var result []*v1alpha1.EnvironmentMapping
	for _, item := range items {
		em, ok := item.(*v1alpha1.EnvironmentMapping)
		if !ok {
			return nil, fmt.Errorf("unexpected item type: %T, should be *v1alpha1.EnvironmentMapping", item)
		}
		result = append(result, em)
	}

	return result, nil
}
