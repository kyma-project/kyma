package access

import (
	emListers "github.com/kyma-project/kyma/components/application-broker/pkg/client/listers/applicationconnector/v1alpha1"
	"github.com/pkg/errors"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
)

// NewApplicationMappingService creates new instance of NewApplicationMappingService
func NewApplicationMappingService(lister emListers.ApplicationMappingLister) *ApplicationMappingService {
	return &ApplicationMappingService{
		lister: lister,
	}
}

// ApplicationMappingService provides methods which checks access based on ApplicationMapping objects.
type ApplicationMappingService struct {
	lister emListers.ApplicationMappingLister
}

// IsApplicationEnabled checks, if ApplicationMapping with given name in the namespace exists
func (c *ApplicationMappingService) IsApplicationEnabled(namespace, name string) (bool, error) {
	_, err := c.lister.ApplicationMappings(namespace).Get(name)
	switch {
	case err == nil:
		return true, nil
	case k8sErrors.IsNotFound(err):
		return false, nil
	default:
		return false, errors.Wrapf(err, "while getting ApplicationMapping %s/%s", namespace, name)
	}
}
