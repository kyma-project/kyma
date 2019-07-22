package broker

import (
	"fmt"

	"context"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ServiceBrokerSyncer provide services to sync the ClusterServiceBroker
type ServiceBrokerSyncer struct {
	clusterBrokerName string
	client            client.Client
	log               logrus.FieldLogger
}

// NewServiceBrokerSyncer allows to sync the ServiceBroker.
func NewServiceBrokerSyncer(cli client.Client, clusterBrokerName string, log logrus.FieldLogger) *ServiceBrokerSyncer {
	return &ServiceBrokerSyncer{
		client:            cli,
		clusterBrokerName: clusterBrokerName,
		log:               log.WithField("service", "clusterservicebroker-syncer"),
	}
}

const maxSyncRetries = 5

// Sync syncs the ServiceBrokers, does not fail if the broker does not exists
func (r *ServiceBrokerSyncer) Sync() error {
	for i := 0; i < maxSyncRetries; i++ {
		broker := &v1beta1.ClusterServiceBroker{}
		err := r.client.Get(context.Background(), types.NamespacedName{Name: r.clusterBrokerName}, broker)
		switch {
		case apiErrors.IsNotFound(err):
			return nil
		case err != nil:
			return errors.Wrapf(err, "while getting ClusterServiceBrokers %s", r.clusterBrokerName)
		}

		// update RelistRequests to trigger the relist
		broker.Spec.RelistRequests = broker.Spec.RelistRequests + 1

		err = r.client.Update(context.Background(), broker)
		switch {
		case err == nil:
			return nil
		case apiErrors.IsConflict(err):
			r.log.Infof("(%d/%d) ClusterServiceBroker %s update conflict occurred.", i, maxSyncRetries, broker.Name)
		case err != nil:
			return errors.Wrapf(err, "while updating ClusterServiceBroker %s", broker.Name)
		}
	}

	return fmt.Errorf("could not sync cluster service broker (%s) after %d retries", r.clusterBrokerName, maxSyncRetries)
}

// SyncServiceBroker syncing the helm-broker ns-broker in the given namespace
func (r *ServiceBrokerSyncer) SyncServiceBroker(namespace string) error {
	for i := 0; i < maxSyncRetries; i++ {
		broker := &v1beta1.ServiceBroker{}
		err := r.client.Get(context.Background(), types.NamespacedName{Name: NamespacedBrokerName, Namespace: namespace}, broker)
		switch {
		case apiErrors.IsNotFound(err):
			return nil
		case err != nil:
			return errors.Wrapf(err, "while getting ClusterServiceBrokers %s", r.clusterBrokerName)
		}

		broker.Spec.RelistRequests = broker.Spec.RelistRequests + 1

		err = r.client.Update(context.Background(), broker)
		switch {
		case err == nil:
			return nil
		case apiErrors.IsConflict(err):
			r.log.Infof("(%d/%d) ServiceBroker %s update conflict occurred.", i, maxSyncRetries, broker.Name)
		case err != nil:
			return errors.Wrapf(err, "while updating ClusterServiceBroker %s", broker.Name)
		}
	}
	return fmt.Errorf("could not sync service broker (%s) after %d retries", NamespacedBrokerName, maxSyncRetries)
}
