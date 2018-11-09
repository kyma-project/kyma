package broker

import (
	"fmt"

	v1beta12 "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterServiceBrokerSync provide services to sync the ClusterServiceBroker
type ClusterServiceBrokerSync struct {
	csbInterface v1beta12.ClusterServiceBrokerInterface
	log          logrus.FieldLogger
}

// NewClusterServiceBrokerSyncer allows to sync the ServiceBroker.
func NewClusterServiceBrokerSyncer(csbInterface v1beta12.ClusterServiceBrokerInterface, log logrus.FieldLogger) *ClusterServiceBrokerSync {
	return &ClusterServiceBrokerSync{
		csbInterface: csbInterface,
		log:          log.WithField("service", "clusterservicebroker-syncer"),
	}
}

// Sync syncs the ServiceBrokers, does not fail if the broker does not exists
func (r *ClusterServiceBrokerSync) Sync(name string, maxSyncRetries int) error {
	r.log.Infof("Trigger Service Catalog to refresh ClusterServiceBroker %s", name)
	for i := 0; i < maxSyncRetries; i++ {
		broker, err := r.csbInterface.Get(name, v1.GetOptions{})
		switch {
		// do not return error if the broker does not exists, the method is dedicated to update
		// ClusterServiceBroker resource if exists. If it is not created yet
		// - it will be created in the future and Service Catalog will call the 'catalog' endpoint soon.
		case apiErrors.IsNotFound(err):
			return nil
		case err != nil:
			return errors.Wrapf(err, "while getting ClusterServiceBrokers %s", name)
		}

		// update RelistRequests to trigger the relist
		broker.Spec.RelistRequests = broker.Spec.RelistRequests + 1

		_, err = r.csbInterface.Update(broker)
		switch {
		case err == nil:
			return nil
		case apiErrors.IsConflict(err):
			r.log.Infof("(%d/%d) ClusterServiceBroker %s update conflict occurred.", i, maxSyncRetries, broker.Name)
		case err != nil:
			return errors.Wrapf(err, "while updating ClusterServiceBroker %s", broker.Name)
		}
	}

	return fmt.Errorf("could not sync cluster service broker (%s) after %d retries", name, maxSyncRetries)
}
