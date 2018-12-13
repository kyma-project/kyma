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

//go:generate mockery -name=applicationFinder -output=automock -outpkg=automock -case=underscore
type applicationFinder interface {
	FindOneByServiceID(id internal.ApplicationServiceID) (*internal.Application, error)
}

// NewMappingExistsProvisionChecker creates new access checker
func NewMappingExistsProvisionChecker(appFinder applicationFinder, appInterface versioned.ApplicationconnectorV1alpha1Interface) *MappingExistsProvisionChecker {
	return &MappingExistsProvisionChecker{
		appInterface: appInterface,
		appFinder:    appFinder,
	}
}

// MappingExistsProvisionChecker is a checker which can wait some time for ApplicationMapping before it forbids provisioning
type MappingExistsProvisionChecker struct {
	appInterface versioned.ApplicationconnectorV1alpha1Interface
	appFinder    applicationFinder
}

// CanProvision checks if service instance can be provisioned in the namespace
func (c *MappingExistsProvisionChecker) CanProvision(serviceID internal.ApplicationServiceID, namespace internal.Namespace, maxWaitTime time.Duration) (CanProvisionOutput, error) {
	app, err := c.appFinder.FindOneByServiceID(serviceID)
	if err != nil {
		return CanProvisionOutput{}, errors.Wrapf(err, "while finding application which contains service [%s]", serviceID)
	}
	if app == nil {
		return CanProvisionOutput{}, fmt.Errorf("cannot find application which contains service serviceID: [%s]", serviceID)

	}
	demandedRemoteEnvName := string(app.Name)

	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return c.appInterface.ApplicationMappings(string(namespace)).List(options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return c.appInterface.ApplicationMappings(string(namespace)).Watch(options)
		},
	}

	_, err = cache.ListWatchUntil(maxWaitTime, lw, func(event watch.Event) (bool, error) {
		deepCopy := event.Object.DeepCopyObject()
		envMapping, ok := deepCopy.(*v1alpha1.ApplicationMapping)
		if !ok {
			return false, fmt.Errorf("cannot covert object [%+v] of type %T to *ApplicationMapping", deepCopy, deepCopy)
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
		return CanProvisionOutput{}, errors.Wrapf(err, "while watching for ApplicationMapping with name: [%s] in the namespace [%s]", demandedRemoteEnvName, namespace)
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
		Reason:  fmt.Sprintf("ApplicationMapping does not exist in the [%s] namespace", ns),
	}
}
