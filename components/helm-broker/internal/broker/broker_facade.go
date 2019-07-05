package broker

import (
	"fmt"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	scbeta "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// NamespacedBrokerName name of the namespaced Service Broker
	NamespacedBrokerName = "helm-broker"
	// BrokerLabelKey key of the namespaced Service Broker label
	BrokerLabelKey = "namespaced-helm-broker"
	// BrokerLabelValue value of the namespaced Service Broker label
	BrokerLabelValue = "true"
)

//go:generate mockery -name=brokerSyncer -output=automock -outpkg=automock -case=underscore
type brokerSyncer interface {
	SyncServiceBroker(namespace string) error
}

// BrokersFacade is responsible for creation k8s objects for namespaced broker
type BrokersFacade struct {
	brokerGetter     scbeta.ServiceBrokersGetter
	workingNamespace string
	serviceName      string
	log              logrus.FieldLogger

	brokerSyncer brokerSyncer
}

// NewBrokersFacade returns facade
func NewBrokersFacade(brokerGetter scbeta.ServiceBrokersGetter, brokerSyncer brokerSyncer,
	workingNamespace, serviceName string, log logrus.FieldLogger) *BrokersFacade {
	return &BrokersFacade{
		brokerGetter:     brokerGetter,
		workingNamespace: workingNamespace,
		brokerSyncer:     brokerSyncer,
		serviceName:      serviceName,
		log:              log.WithField("service", "nsbroker-facade"),
	}
}

// Create creates ServiceBroker. Errors don't stop execution of method. AlreadyExist errors are ignored.
func (f *BrokersFacade) Create(destinationNs string) error {
	var resultErr *multierror.Error

	svcURL := fmt.Sprintf("http://%s.%s.svc.cluster.local", f.serviceName, f.workingNamespace)
	_, err := f.createServiceBroker(svcURL, destinationNs)
	if err != nil {
		resultErr = multierror.Append(resultErr, err)
		f.log.Warnf("Creation of namespaced-broker for namespace [%s] results in error: [%s]. AlreadyExist errors will be ignored.", destinationNs, err)
	}

	f.log.Infof("Triggering Service Catalog to do a sync with a broker in %s namespace", destinationNs)
	err = f.brokerSyncer.SyncServiceBroker(destinationNs)
	if err != nil {
		f.log.Warnf("Failed to sync a broker in the namespace %s: %s", destinationNs, err.Error())
	}

	resultErr = filterOutMultiError(resultErr, ignoreAlreadyExist)

	if resultErr == nil {
		return nil
	}
	return resultErr
}

// createServiceBroker returns just created or existing ServiceBroker
func (f *BrokersFacade) createServiceBroker(svcURL, namespace string) (*v1beta1.ServiceBroker, error) {
	url := fmt.Sprintf("%s/ns/%s", svcURL, namespace)
	broker := &v1beta1.ServiceBroker{
		ObjectMeta: metav1.ObjectMeta{
			Name:      NamespacedBrokerName,
			Namespace: namespace,
			Labels: map[string]string{
				BrokerLabelKey: BrokerLabelValue,
			},
		},
		Spec: v1beta1.ServiceBrokerSpec{
			CommonServiceBrokerSpec: v1beta1.CommonServiceBrokerSpec{
				URL: url,
			},
		},
	}

	createdBroker, err := f.brokerGetter.ServiceBrokers(namespace).Create(broker)
	if k8serrors.IsAlreadyExists(err) {
		f.log.Infof("ServiceBroker for namespace [%s] already exist. Attempt to get resource.", namespace)
		createdBroker, err = f.brokerGetter.ServiceBrokers(namespace).Get(NamespacedBrokerName, metav1.GetOptions{})
		return createdBroker, err
	}

	return createdBroker, err
}

// Delete removes ServiceBroker and BrokersFacade. Errors don't stop execution of method. NotFound errors are ignored.
func (f *BrokersFacade) Delete(destinationNs string) error {
	err := f.brokerGetter.ServiceBrokers(destinationNs).Delete(NamespacedBrokerName, nil)
	switch {
	case k8serrors.IsNotFound(err):
		return nil
	case err != nil:
		f.log.Warnf("Deletion of namespaced-broker for namespace [%s] results in error: [%s].", destinationNs, err)
	}
	return err

}

// Exist check if ServiceBroker exists.
func (f *BrokersFacade) Exist(destinationNs string) (bool, error) {
	_, err := f.brokerGetter.ServiceBrokers(destinationNs).Get(NamespacedBrokerName, metav1.GetOptions{})
	switch {
	case k8serrors.IsNotFound(err):
		return false, nil
	case err != nil:
		return false, errors.Wrapf(err, "while checking if ServiceBroker [%s] exists in the namespace [%s]", NamespacedBrokerName, destinationNs)
	}

	return true, nil
}

func filterOutMultiError(merr *multierror.Error, predicate func(err error) bool) *multierror.Error {
	if merr == nil {
		return nil
	}
	var out *multierror.Error
	for _, wrapped := range merr.Errors {
		if predicate(wrapped) {
			out = multierror.Append(out, wrapped)
		}
	}
	return out

}

func ignoreAlreadyExist(err error) bool {
	return !k8serrors.IsAlreadyExists(err)
}

func ignoreIsNotFound(err error) bool {
	return !k8serrors.IsNotFound(err)
}
