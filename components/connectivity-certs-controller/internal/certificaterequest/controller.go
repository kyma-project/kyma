package certificaterequest

import (
	"context"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"

	"github.com/kyma-project/kyma/components/connectivity-certs-controller/internal/certificates"
	"github.com/kyma-project/kyma/components/connectivity-certs-controller/internal/connectorservice"
	"github.com/kyma-project/kyma/components/connectivity-certs-controller/pkg/apis/applicationconnector/v1alpha1"
	log "github.com/sirupsen/logrus"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type Client interface {
	Get(ctx context.Context, key client.ObjectKey, obj runtime.Object) error
	Update(ctx context.Context, obj runtime.Object) error
	Delete(ctx context.Context, obj runtime.Object, opts ...client.DeleteOptionFunc) error
}

type CentralConnectionManager interface {
	Create(*v1alpha1.CentralConnection) (*v1alpha1.CentralConnection, error)
}

type Controller struct {
	certificateRequestClient Client
	connectorClient          connectorservice.Client
	certificatePreserver     certificates.Preserver
	connectionClient         CentralConnectionManager
}

func InitCertificatesRequestController(mgr manager.Manager, appName string, connectorClient connectorservice.Client, certPreserver certificates.Preserver, connectionManager CentralConnectionManager) error {
	return startController(appName, mgr, connectorClient, certPreserver, connectionManager)
}

func startController(appName string, mgr manager.Manager, connectorClient connectorservice.Client, certPreserver certificates.Preserver, connectionManager CentralConnectionManager) error {
	certRequestController := newCertificatesRequestController(mgr.GetClient(), connectorClient, certPreserver, connectionManager)

	c, err := controller.New(appName, mgr, controller.Options{Reconciler: certRequestController})
	if err != nil {
		return err
	}

	return c.Watch(&source.Kind{Type: &v1alpha1.CertificateRequest{}}, &handler.EnqueueRequestForObject{})
}

func newCertificatesRequestController(client Client, connectorClient connectorservice.Client, certPreserver certificates.Preserver, connectionManager CentralConnectionManager) *Controller {
	return &Controller{
		certificateRequestClient: client,
		connectorClient:          connectorClient,
		certificatePreserver:     certPreserver,
		connectionClient:         connectionManager,
	}
}

func (c *Controller) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	instance := &v1alpha1.CertificateRequest{}

	log.Infof("Processing %s request", request.Name)

	err := c.certificateRequestClient.Get(context.Background(), request.NamespacedName, instance)
	if err != nil {
		return c.handleErrorWhileGettingInstance(err, request)
	}

	if instance.Status.Error != "" {
		log.Infof("Certificate Request %s has an error status, certificate will not be fetched", instance.Name)
		return reconcile.Result{}, nil
	}

	establishedConnection, err := c.connectorClient.ConnectToCentralConnector(instance.Spec.CSRInfoURL)
	if err != nil {
		log.Errorf("Error while requesting certificates from Connector Service: %s", err.Error())
		return reconcile.Result{}, c.setRequestErrorStatus(instance, err)
	}
	log.Infoln("Certificates fetched successfully")

	err = c.manageResources(instance.Name, establishedConnection)
	if err != nil {
		return reconcile.Result{}, c.setRequestErrorStatus(instance, err)
	}

	err = c.certificateRequestClient.Delete(context.Background(), instance)
	if err != nil {
		return reconcile.Result{}, err
	}
	log.Infof("CertificatesRequest %s successfully deleted", instance.Name)

	return reconcile.Result{}, nil
}

func (c *Controller) manageResources(connectionName string, connection connectorservice.EstablishedConnection) error {
	masterConnection := &v1alpha1.CentralConnection{
		ObjectMeta: v1.ObjectMeta{Name: connectionName}, // TODO - figure out some naming - same as request??
		Spec: v1alpha1.CentralConnectionSpec{
			ManagementInfoURL: connection.ManagementInfoURL,
			// TODO - secret names?
		},
	}

	_, err := c.connectionClient.Create(masterConnection)
	if err != nil {
		log.Errorf("Error while creating Central Connection resource: %s", err.Error())
	}

	err = c.certificatePreserver.PreserveCertificates(connection.Certificates)
	if err != nil {
		log.Errorf("Error while saving certificates to secrets: %s", err.Error())
		return err
	}
	log.Infoln("Certificates saved successfully")

	return nil
}

func (c *Controller) handleErrorWhileGettingInstance(err error, request reconcile.Request) (reconcile.Result, error) {
	if k8sErrors.IsNotFound(err) {
		log.Infof("Request %s has been deleted", request.Name)
		return reconcile.Result{}, nil
	}
	log.Errorf("Error while getting instance, %s", err.Error())
	return reconcile.Result{}, err
}

func (c *Controller) setRequestErrorStatus(instance *v1alpha1.CertificateRequest, statusError error) error {
	instance.Status = v1alpha1.CertificateRequestStatus{
		Error: statusError.Error(),
	}

	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		return c.certificateRequestClient.Update(context.Background(), instance)
	})
	if err != nil {
		return err
	}

	return nil
}
