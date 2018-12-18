package access

import (
	emListers "github.com/kyma-project/kyma/components/application-broker/pkg/client/listers/applicationconnector/v1alpha1"
	"github.com/pkg/errors"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
)

// NewEnvironmentMappingService creates new instance of NewEnvironmentMappingService
func NewEnvironmentMappingService(lister emListers.EnvironmentMappingLister) *EnvironmentMappingService {
	return &EnvironmentMappingService{
		lister: lister,
	}
}

// EnvironmentMappingService provides methods which checks access based on EnvironmentMapping objects.
type EnvironmentMappingService struct {
	lister emListers.EnvironmentMappingLister
}

// IsRemoteEnvironmentEnabled checks, if EnvironmentMapping with given name in the namespace exists
func (c *EnvironmentMappingService) IsRemoteEnvironmentEnabled(namespace, name string) (bool, error) {
	_, err := c.lister.EnvironmentMappings(namespace).Get(name)
	switch {
	case err == nil:
		return true, nil
	case k8sErrors.IsNotFound(err):
		return false, nil
	default:
		return false, errors.Wrapf(err, "while getting EnvironmentMapping %s/%s", namespace, name)
	}
}
