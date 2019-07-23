package broker

import (
	"fmt"

	"context"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type clusterBrokerSyncer interface {
	Sync() error
}

// ClusterBrokersFacade is responsible for creation k8s objects for namespaced broker
type ClusterBrokersFacade struct {
	client            client.Client
	workingNamespace  string
	serviceName       string
	log               logrus.FieldLogger
	clusterBrokerName string

	clusterBrokerSyncer clusterBrokerSyncer
}

// NewClusterBrokersFacade returns facade
func NewClusterBrokersFacade(client client.Client, clusterBrokerSyncer clusterBrokerSyncer, workingNamespace, serviceName, clusterBrokerName string, log logrus.FieldLogger) *ClusterBrokersFacade {
	return &ClusterBrokersFacade{
		client:              client,
		workingNamespace:    workingNamespace,
		clusterBrokerSyncer: clusterBrokerSyncer,
		clusterBrokerName:   clusterBrokerName,
		serviceName:         serviceName,
		log:                 log.WithField("service", "cluster-broker-facade"),
	}
}

// Create creates ClusterServiceBroker. Errors don't stop execution of method. AlreadyExist errors are ignored.
func (f *ClusterBrokersFacade) Create() error {
	var resultErr *multierror.Error

	svcURL := fmt.Sprintf("http://%s.%s.svc.cluster.local", f.serviceName, f.workingNamespace)
	_, err := f.createClusterServiceBroker(svcURL)
	if err != nil {
		resultErr = multierror.Append(resultErr, err)
		f.log.Warnf("Creation of ClusterServiceBroker %s results in error: [%s]. AlreadyExist errors will be ignored.", f.clusterBrokerName, err)
	}

	err = f.clusterBrokerSyncer.Sync()
	if err != nil {
		f.log.Warnf("Failed to sync a broker %s : %v", f.clusterBrokerName, err.Error())
	}

	resultErr = filterOutMultiError(resultErr, ignoreAlreadyExist)

	if resultErr == nil {
		return nil
	}
	return resultErr
}

// Delete removes ClusterServiceBroker and BrokersFacade. Errors don't stop execution of method. NotFound errors are ignored.
func (f *ClusterBrokersFacade) Delete() error {
	csb := &v1beta1.ClusterServiceBroker{
		ObjectMeta: metav1.ObjectMeta{
			Name: f.clusterBrokerName,
		},
	}

	err := f.client.Delete(context.Background(), csb)
	switch {
	case k8serrors.IsNotFound(err):
		return nil
	case err != nil:
		f.log.Warnf("Deletion of ClusterServiceBroker %s results in error: [%s].", f.clusterBrokerName, err)
	}
	return err

}

// Exist check if ClusterServiceBroker exists.
func (f *ClusterBrokersFacade) Exist() (bool, error) {
	err := f.client.Get(context.Background(), types.NamespacedName{Name: f.clusterBrokerName}, &v1beta1.ClusterServiceBroker{})
	switch {
	case k8serrors.IsNotFound(err):
		return false, nil
	case err != nil:
		return false, errors.Wrapf(err, "while checking if ClusterServiceBroker [%s] exists", f.clusterBrokerName)
	}

	return true, nil
}

// createServiceBroker returns just created or existing ClusterServiceBroker
func (f *ClusterBrokersFacade) createClusterServiceBroker(svcURL string) (*v1beta1.ClusterServiceBroker, error) {
	url := fmt.Sprintf("%s/cluster", svcURL)
	broker := &v1beta1.ClusterServiceBroker{
		ObjectMeta: metav1.ObjectMeta{
			Name: f.clusterBrokerName,
		},
		Spec: v1beta1.ClusterServiceBrokerSpec{
			CommonServiceBrokerSpec: v1beta1.CommonServiceBrokerSpec{
				URL: url,
			},
		},
	}

	err := f.client.Create(context.Background(), broker)
	if k8serrors.IsAlreadyExists(err) {
		f.log.Infof("ClusterServiceBroker [%s] already exist. Attempt to get resource.", broker.Name)
		createdBroker := &v1beta1.ClusterServiceBroker{}
		err = f.client.Get(context.Background(), types.NamespacedName{Name: f.clusterBrokerName}, createdBroker)
		return createdBroker, err
	}

	return broker, err
}
