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

type clusterBrokerSyncer interface {
	Sync(name string) error
}

// ClusterBrokersFacade is responsible for creation k8s objects for namespaced broker
type ClusterBrokersFacade struct {
	clusterBrokerGetter scbeta.ClusterServiceBrokersGetter
	workingNamespace    string
	serviceName         string
	log                 logrus.FieldLogger

	clusterBrokerSyncer clusterBrokerSyncer
}

// NewClusterBrokersFacade returns facade
func NewClusterBrokersFacade(clusterBrokerGetter scbeta.ClusterServiceBrokersGetter, clusterBrokerSyncer clusterBrokerSyncer,
	workingNamespace, serviceName string, log logrus.FieldLogger) *ClusterBrokersFacade {
	return &ClusterBrokersFacade{
		clusterBrokerGetter: clusterBrokerGetter,
		workingNamespace:    workingNamespace,
		clusterBrokerSyncer: clusterBrokerSyncer,
		serviceName:         serviceName,
		log:                 log.WithField("service", "cluster-broker-facade"),
	}
}

const brokerName = "helm-broker"

// Create creates ClusterServiceBroker. Errors don't stop execution of method. AlreadyExist errors are ignored.
func (f *ClusterBrokersFacade) Create() error {
	var resultErr *multierror.Error

	svcURL := fmt.Sprintf("http://%s.%s.svc.cluster.local", f.serviceName, f.workingNamespace)
	_, err := f.createClusterServiceBroker(svcURL)
	if err != nil {
		resultErr = multierror.Append(resultErr, err)
		f.log.Warnf("Creation of ClusterServiceBroker %s results in error: [%s]. AlreadyExist errors will be ignored.", brokerName, err)
	}

	f.log.Infof("Triggering Service Catalog to do a sync with a ClusterServiceBroker %s", brokerName)
	err = f.clusterBrokerSyncer.Sync(brokerName)
	if err != nil {
		f.log.Warnf("Failed to sync a broker %s : %v", brokerName, err.Error())
	}

	resultErr = filterOutMultiError(resultErr, ignoreAlreadyExist)

	if resultErr == nil {
		return nil
	}
	return resultErr
}

// Delete removes ClusterServiceBroker and BrokersFacade. Errors don't stop execution of method. NotFound errors are ignored.
func (f *ClusterBrokersFacade) Delete() error {
	err := f.clusterBrokerGetter.ClusterServiceBrokers().Delete(brokerName, nil)
	switch {
	case k8serrors.IsNotFound(err):
		return nil
	case err != nil:
		f.log.Warnf("Deletion of ClusterServiceBroker %s results in error: [%s].", brokerName, err)
	}
	return err

}

// Exist check if ClusterServiceBroker exists.
func (f *ClusterBrokersFacade) Exist() (bool, error) {
	_, err := f.clusterBrokerGetter.ClusterServiceBrokers().Get(brokerName, metav1.GetOptions{})
	switch {
	case k8serrors.IsNotFound(err):
		return false, nil
	case err != nil:
		return false, errors.Wrapf(err, "while checking if ServiceBroker [%s] exists in the namespace [%s]", brokerName)
	}

	return true, nil
}

// createServiceBroker returns just created or existing ClusterServiceBroker
func (f *ClusterBrokersFacade) createClusterServiceBroker(svcURL string) (*v1beta1.ClusterServiceBroker, error) {
	url := fmt.Sprintf("%s/cluster", svcURL)
	broker := &v1beta1.ClusterServiceBroker{
		ObjectMeta: metav1.ObjectMeta{
			Name: brokerName,
		},
		Spec: v1beta1.ClusterServiceBrokerSpec{
			CommonServiceBrokerSpec: v1beta1.CommonServiceBrokerSpec{
				URL: url,
			},
		},
	}

	createdBroker, err := f.clusterBrokerGetter.ClusterServiceBrokers().Create(broker)
	if k8serrors.IsAlreadyExists(err) {
		f.log.Infof("ClusterServiceBroker [%s] already exist. Attempt to get resource.", broker.Name)
		createdBroker, err = f.clusterBrokerGetter.ClusterServiceBrokers().Get(brokerName, metav1.GetOptions{})
		return createdBroker, err
	}

	return createdBroker, err
}
