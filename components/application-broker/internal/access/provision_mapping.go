package access

import (
	"fmt"
	"time"

	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	versioned "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

//go:generate mockery -name=remoteEnvironmentFinder -output=automock -outpkg=automock -case=underscore
type remoteEnvironmentFinder interface {
	FindOneByServiceID(id internal.RemoteServiceID) (*internal.RemoteEnvironment, error)
}

// NewMappingExistsProvisionChecker creates new access checker
func NewMappingExistsProvisionChecker(reFinder remoteEnvironmentFinder, reInterface versioned.ApplicationconnectorV1alpha1Interface) *MappingExistsProvisionChecker {
	return &MappingExistsProvisionChecker{
		reInterface: reInterface,
		reFinder:    reFinder,
	}
}

// MappingExistsProvisionChecker is a checker which can wait some time for EnvironmentMapping before it forbids provisioning
type MappingExistsProvisionChecker struct {
	reInterface versioned.ApplicationconnectorV1alpha1Interface
	reFinder    remoteEnvironmentFinder
}

// CanProvision checks if service instance can be provisioned in the namespace
func (c *MappingExistsProvisionChecker) CanProvision(serviceID internal.RemoteServiceID, namespace internal.Namespace, maxWaitTime time.Duration) (CanProvisionOutput, error) {
	re, err := c.reFinder.FindOneByServiceID(serviceID)
	if err != nil {
		return CanProvisionOutput{}, errors.Wrapf(err, "while finding remote environment which contains service [%s]", serviceID)
	}
	if re == nil {
		return CanProvisionOutput{}, fmt.Errorf("cannot find remote environment which contains service serviceID: [%s]", serviceID)

	}
	demandedRemoteEnvName := string(re.Name)

	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return c.reInterface.EnvironmentMappings(string(namespace)).List(options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return c.reInterface.EnvironmentMappings(string(namespace)).Watch(options)
		},
	}

	_, err = cache.ListWatchUntil(maxWaitTime, lw, func(event watch.Event) (bool, error) {
		deepCopy := event.Object.DeepCopyObject()
		envMapping, ok := deepCopy.(*v1alpha1.EnvironmentMapping)
		if !ok {
			return false, fmt.Errorf("cannot covert object [%+v] of type %T to *EnvironmentMapping", deepCopy, deepCopy)
		}

		if envMapping.Name == demandedRemoteEnvName {
			return true, nil
		}
		return false, nil
	})

	switch err {
	case nil:
		return c.responseAllow(), nil
	case wait.ErrWaitTimeout:
		return c.responseDeny(namespace), nil
	default:
		return CanProvisionOutput{}, errors.Wrapf(err, "while watching for EnvironmentMapping with name: [%s] in the namespace [%s]", demandedRemoteEnvName, namespace)
	}

}

func (c *MappingExistsProvisionChecker) responseAllow() CanProvisionOutput {
	return CanProvisionOutput{
		Allowed: true,
	}
}

func (c *MappingExistsProvisionChecker) responseDeny(ns internal.Namespace) CanProvisionOutput {
	return CanProvisionOutput{
		Allowed: false,
		Reason:  fmt.Sprintf("EnvironmentMapping does not exist in the [%s] namespace", ns),
	}
}
