package access

import (
	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	emListers "github.com/kyma-project/kyma/components/application-broker/pkg/client/listers/applicationconnector/v1alpha1"
	"github.com/pkg/errors"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
)

// ServiceEnabledChecker provides a method to check if a Service is enabled.
type ServiceEnabledChecker interface {
	// IsServiceEnabled returns true if the service is enabled
	IsServiceEnabled(svc internal.Service) bool
}

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

// NewServiceChecker creates a ServiceEnabledChecker which is able to check if a service of an application
// is enabled in the namespace.
func (c *ApplicationMappingService) NewServiceChecker(namespace, name string) (ServiceEnabledChecker, error) {
	am, err := c.lister.ApplicationMappings(namespace).Get(name)
	switch {
	case err == nil:
		return newApplicationServiceChecker(am), nil
	case k8sErrors.IsNotFound(err):
		return &allServicesDisabled{}, nil
	default:
		return nil, errors.Wrapf(err, "while getting ApplicationMapping %s/%s", namespace, name)
	}
}

func newApplicationServiceChecker(am *v1alpha1.ApplicationMapping) ServiceEnabledChecker {
	if am.IsAllApplicationServicesEnabled() {
		return &allServicesEnabled{}
	}
	serviceIDs := make(map[internal.ApplicationServiceID]struct{})

	for _, svc := range am.Spec.Services {
		serviceIDs[internal.ApplicationServiceID(svc.ID)] = struct{}{}
	}

	return &ApplicationServiceChecker{
		serviceIDs: serviceIDs,
	}
}

// ApplicationServiceChecker is a default ServiceEnabledChecker implementation
type ApplicationServiceChecker struct {
	serviceIDs map[internal.ApplicationServiceID]struct{}
}

// IsServiceEnabled returns true if the service is enabled
func (c *ApplicationServiceChecker) IsServiceEnabled(svc internal.Service) bool {
	_, exists := c.serviceIDs[svc.ID]
	return exists
}

type allServicesDisabled struct {
}

// IsServiceEnabled implements ServiceEnabledChecker always returns false
func (c *allServicesDisabled) IsServiceEnabled(svc internal.Service) bool {
	return false
}

type allServicesEnabled struct {
}

// IsServiceEnabled implements ServiceEnabledChecker always returns true
func (c *allServicesEnabled) IsServiceEnabled(svc internal.Service) bool {
	return true
}
