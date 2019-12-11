package syncer

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
	v1beta12 "github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/application-broker/internal/nsbroker"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ServiceBrokerSync provide services to sync the ServiceBrokers
type ServiceBrokerSync struct {
	serviceBrokerGetter v1beta12.ServiceBrokersGetter
	log                 logrus.FieldLogger
}

type brokerInfo struct {
	name      string
	namespace string
}

// NewServiceBrokerSyncer allows to sync the ServiceBrokers
func NewServiceBrokerSyncer(serviceBrokerGetter v1beta12.ServiceBrokersGetter) *ServiceBrokerSync {
	return &ServiceBrokerSync{
		serviceBrokerGetter: serviceBrokerGetter,
		log:                 logrus.WithField("service", "syncer:ns-broker-syncer"),
	}
}

// Sync syncs the ServiceBrokers
func (r *ServiceBrokerSync) Sync(maxSyncRetries int) error {
	labelSelector := fmt.Sprintf("%s=%s", nsbroker.BrokerLabelKey, nsbroker.BrokerLabelValue)
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

// SyncBroker syncing the default AB ns-broker in the given namespace
func (r *ServiceBrokerSync) SyncBroker(namespace string) error {
	brokerClient := r.serviceBrokerGetter.ServiceBrokers(namespace)
	name := nsbroker.NamespacedBrokerName

	for i := 0; i < maxSyncRetries; i++ {
		broker, err := brokerClient.Get(name, v1.GetOptions{})
		if err != nil {
			return errors.Wrapf(err, "while getting ServiceBroker %q [namespace: %s]", name, namespace)
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
	return fmt.Errorf("could not sync service broker (%s) after %d retries", nsbroker.NamespacedBrokerName, maxSyncRetries)
}
