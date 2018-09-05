package syncer

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
	v1beta12 "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
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
func (r *ServiceBrokerSync) Sync(labelSelector string, maxSyncRetries int) error {
	brokersList, err := r.serviceBrokerGetter.ServiceBrokers(v1.NamespaceAll).List(v1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return errors.Wrapf(err, "while listing ServiceBrokers [labelSelector: %s]", labelSelector)
	}

	brokersInfo := make([]brokerInfo, 0)
	for _, broker := range brokersList.Items {
		brokersInfo = append(brokersInfo, brokerInfo{
			name:      broker.Name,
			namespace: broker.Namespace,
		})
	}

	var resultErr *multierror.Error
	for _, broker := range brokersInfo {
		for i := 0; i < maxSyncRetries; i++ {
			retrievedBroker, err := r.serviceBrokerGetter.ServiceBrokers(broker.namespace).Get(broker.name, v1.GetOptions{})
			if err != nil {
				resultErr = multierror.Append(resultErr, errors.Wrapf(err, "while getting ServiceBroker %q [namespace: %s]", broker.name, broker.namespace))
				if i == maxSyncRetries-1 {
					resultErr = multierror.Append(resultErr, fmt.Errorf("could not sync ServiceBroker %q [namespace: %s], after %d tries", broker.name, broker.namespace, maxSyncRetries))
				}
				continue
			}

			retrievedBroker.Spec.RelistRequests = retrievedBroker.Spec.RelistRequests + 1
			_, err = r.serviceBrokerGetter.ServiceBrokers(broker.namespace).Update(retrievedBroker)
			if err == nil {
				r.log.Infof("Relist request for ServiceBroker %q [namespace: %s] fulfilled", broker.name, broker.namespace)
				break
			}
			if !apiErrors.IsConflict(err) {
				resultErr = multierror.Append(resultErr, errors.Wrapf(err, "while updating ServiceBroker %q [namespace: %s]", broker.name, broker.namespace))
			}
			if i == maxSyncRetries-1 {
				resultErr = multierror.Append(resultErr, fmt.Errorf("could not sync ServiceBroker %q [namespace: %s], after %d tries", broker.name, broker.namespace, maxSyncRetries))
			}
		}
	}
	if resultErr == nil {
		return nil
	}
	return resultErr
}
