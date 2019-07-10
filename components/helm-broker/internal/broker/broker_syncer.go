package broker

import (
	"fmt"

	multierror "github.com/hashicorp/go-multierror"
	v1beta12 "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ServiceBrokerSyncer provide services to sync the ClusterServiceBroker
type ServiceBrokerSyncer struct {
	clusterServiceBrokersGetter v1beta12.ClusterServiceBrokersGetter
	serviceBrokerGetter         v1beta12.ServiceBrokersGetter
	clusterBrokerName           string
	log                         logrus.FieldLogger
}

// NewServiceBrokerSyncer allows to sync the ServiceBroker.
func NewServiceBrokerSyncer(clusterServiceBrokersGetter v1beta12.ClusterServiceBrokersGetter, serviceBrokerGetter v1beta12.ServiceBrokersGetter, clusterBrokerName string, log logrus.FieldLogger) *ServiceBrokerSyncer {
	return &ServiceBrokerSyncer{
		clusterServiceBrokersGetter: clusterServiceBrokersGetter,
		clusterBrokerName:           clusterBrokerName,
		serviceBrokerGetter:         serviceBrokerGetter,
		log:                         log.WithField("service", "clusterservicebroker-syncer"),
	}
}

const maxSyncRetries = 5

// Sync syncs the ServiceBrokers, does not fail if the broker does not exists
func (r *ServiceBrokerSyncer) Sync() error {
	r.log.Infof("Trigger Service Catalog to refresh ClusterServiceBroker %s", r.clusterBrokerName)
	for i := 0; i < maxSyncRetries; i++ {
		broker, err := r.clusterServiceBrokersGetter.ClusterServiceBrokers().Get(r.clusterBrokerName, v1.GetOptions{})
		switch {
		// do not return error if the broker does not exists, the method is dedicated to update
		// ClusterServiceBroker resource if exists. If it is not created yet
		// - it will be created in the future and Service Catalog will call the 'catalog' endpoint soon.
		case apiErrors.IsNotFound(err):
			return nil
		case err != nil:
			return errors.Wrapf(err, "while getting ClusterServiceBrokers %s", r.clusterBrokerName)
		}

		// update RelistRequests to trigger the relist
		broker.Spec.RelistRequests = broker.Spec.RelistRequests + 1

		_, err = r.clusterServiceBrokersGetter.ClusterServiceBrokers().Update(broker)
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
	brokerClient := r.serviceBrokerGetter.ServiceBrokers(namespace)

	for i := 0; i < maxSyncRetries; i++ {
		broker, err := brokerClient.Get(NamespacedBrokerName, v1.GetOptions{})
		if err != nil {
			return errors.Wrapf(err, "while getting ServiceBroker %q [namespace: %s]", NamespacedBrokerName, namespace)
		}

		broker.Spec.RelistRequests = broker.Spec.RelistRequests + 1

		_, err = brokerClient.Update(broker)
		if err == nil {
			return nil
		}
		if !apiErrors.IsConflict(err) {
			return fmt.Errorf("could not sync service broker (%s)", err)
		}
	}
	return fmt.Errorf("could not sync service broker (%s) after %d retries", NamespacedBrokerName, maxSyncRetries)
}

// SyncServiceBrokers syncs the ServiceBrokers
func (r *ServiceBrokerSyncer) SyncServiceBrokers() error {
	labelSelector := fmt.Sprintf("%s=%s", BrokerLabelKey, BrokerLabelValue)
	brokersList, err := r.serviceBrokerGetter.ServiceBrokers(v1.NamespaceAll).List(v1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return errors.Wrapf(err, "while listing ServiceBrokers [labelSelector: %s]", labelSelector)
	}
	r.log.Infof("There are %d ServiceBroker(s) with label: %s", len(brokersList.Items), labelSelector)

	var resultErr *multierror.Error
	for _, broker := range brokersList.Items {
		for i := 0; i < maxSyncRetries; i++ {
			retrievedBroker, err := r.serviceBrokerGetter.ServiceBrokers(broker.Namespace).Get(broker.Name, v1.GetOptions{})
			if err != nil {
				resultErr = multierror.Append(resultErr, errors.Wrapf(err, "while getting ServiceBroker %q [namespace: %s]", broker.Name, broker.Namespace))
				if i == maxSyncRetries-1 {
					resultErr = multierror.Append(resultErr, fmt.Errorf("could not sync ServiceBroker %q [namespace: %s], after %d tries", broker.Name, broker.Namespace, maxSyncRetries))
				}
				continue
			}

			retrievedBroker.Spec.RelistRequests = retrievedBroker.Spec.RelistRequests + 1
			_, err = r.serviceBrokerGetter.ServiceBrokers(broker.Namespace).Update(retrievedBroker)
			if err == nil {
				r.log.Infof("Relist request for ServiceBroker %q [namespace: %s] fulfilled", broker.Name, broker.Namespace)
				break
			}
			if !apiErrors.IsConflict(err) {
				resultErr = multierror.Append(resultErr, errors.Wrapf(err, "while updating ServiceBroker %q [namespace: %s]", broker.Name, broker.Namespace))
			}
			if i == maxSyncRetries-1 {
				resultErr = multierror.Append(resultErr, fmt.Errorf("could not sync ServiceBroker %q [namespace: %s], after %d tries", broker.Name, broker.Namespace, maxSyncRetries))
			}
		}
	}
	if resultErr == nil {
		return nil
	}
	return resultErr
}
