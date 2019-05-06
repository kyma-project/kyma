package certificaterequest

import (
	"context"
	"time"

	"github.com/kyma-project/kyma/components/connectivity-certs-controller/internal/centralconnection"

	"k8s.io/apimachinery/pkg/types"

	"k8s.io/client-go/util/retry"

	"github.com/kyma-project/kyma/components/connectivity-certs-controller/internal/certificates"
	"github.com/kyma-project/kyma/components/connectivity-certs-controller/internal/connectorservice"
	"github.com/kyma-project/kyma/components/connectivity-certs-controller/pkg/apis/applicationconnector/v1alpha1"
	log "github.com/sirupsen/logrus"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

type Controller struct {
	certificateRequestClient Client
	initialConnectionClient  connectorservice.InitialConnectionClient
	certificatePreserver     certificates.Preserver
	connectionClient         centralconnection.Client
	logger                   *log.Entry
}

func InitCertificatesRequestController(mgr manager.Manager, appName string, connectorClient connectorservice.InitialConnectionClient, certPreserver certificates.Preserver, connectionClient centralconnection.Client) error {
	return startController(appName, mgr, connectorClient, certPreserver, connectionClient)
}

func startController(appName string, mgr manager.Manager, initialConnectionClient connectorservice.InitialConnectionClient, certPreserver certificates.Preserver, connectionClient centralconnection.Client) error {
	certRequestController := newCertificatesRequestController(mgr.GetClient(), initialConnectionClient, certPreserver, connectionClient)

	c, err := controller.New(appName, mgr, controller.Options{Reconciler: certRequestController})
	if err != nil {
		return err
	}

	return c.Watch(&source.Kind{Type: &v1alpha1.CertificateRequest{}}, &handler.EnqueueRequestForObject{})
}

func newCertificatesRequestController(client Client, connectorClient connectorservice.InitialConnectionClient, certPreserver certificates.Preserver, connectionManager centralconnection.Client) *Controller {
	return &Controller{
		certificateRequestClient: client,
		initialConnectionClient:  connectorClient,
		certificatePreserver:     certPreserver,
		connectionClient:         connectionManager,
		logger:                   log.WithField("Controller", "Certificate Request"),
	}
}

func (c *Controller) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	instance := &v1alpha1.CertificateRequest{}

	c.logger.Infof("Processing %s request", request.Name)

	err := c.certificateRequestClient.Get(context.Background(), request.NamespacedName, instance)
	if err != nil {
		return c.handleErrorWhileGettingInstance(err, request)
	}

	if instance.Status.Error != "" {
		c.logger.Infof("Certificate Request %s has an error status, certificate will not be fetched", instance.Name)
		return reconcile.Result{}, nil
	}

	c.logger.Infof("Establishing connection with Connector Service for %s Cert Request...", instance.Name)
	establishedConnection, err := c.initialConnectionClient.Establish(instance.Spec.CSRInfoURL)
	if err != nil {
		c.logger.Errorf("Error while requesting certificates from Connector Service: %s", err.Error())
		return reconcile.Result{}, c.setRequestErrorStatus(instance, err)
	}
	c.logger.Infof("Certificates for %s Cert Request fetched successfully", instance.Name)

	err = c.manageResources(instance.Name, establishedConnection)
	if err != nil {
		return reconcile.Result{}, c.setRequestErrorStatus(instance, err)
	}

	err = c.certificateRequestClient.Delete(context.Background(), instance)
	if err != nil {
		return reconcile.Result{}, err
	}
	c.logger.Infof("CertificatesRequest %s successfully deleted", instance.Name)

	return reconcile.Result{}, nil
}

func (c *Controller) manageResources(connectionName string, connection connectorservice.EstablishedConnection) error {
	err := c.certificatePreserver.PreserveCertificates(connection.Certificates)
	if err != nil {
		c.logger.Errorf("Error while saving certificates to secrets: %s", err.Error())
		return err
	}
	c.logger.Infoln("Certificates saved successfully")

	centralConnectionSpec := v1alpha1.CentralConnectionSpec{
		ManagementInfoURL: connection.ManagementInfoURL,
		EstablishedAt:     metav1.NewTime(time.Now()),
	}

	_, err = c.connectionClient.Upsert(connectionName, centralConnectionSpec)
	if err != nil {
		c.logger.Errorf("Error while upserting Central Connection resource: %s", err.Error())
		return err
	}

	return nil
}

func (c *Controller) handleErrorWhileGettingInstance(err error, request reconcile.Request) (reconcile.Result, error) {
	if k8sErrors.IsNotFound(err) {
		c.logger.Infof("Request %s has been deleted", request.Name)
		return reconcile.Result{}, nil
	}
	c.logger.Errorf("Error while getting instance, %s", err.Error())
	return reconcile.Result{}, err
}

func (c *Controller) setRequestErrorStatus(instance *v1alpha1.CertificateRequest, statusError error) error {
	name := types.NamespacedName{
		Name:      instance.Name,
		Namespace: instance.Namespace,
	}

	upToDateInstance := &v1alpha1.CertificateRequest{}

	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		err := c.certificateRequestClient.Get(context.Background(), name, upToDateInstance)
		if err != nil {
			c.logger.Errorf("Failed to get up to date resource: %s", err.Error())
			return err
		}

		upToDateInstance.Status = v1alpha1.CertificateRequestStatus{
			Error: statusError.Error(),
		}

		return c.certificateRequestClient.Update(context.Background(), upToDateInstance)
	})
}
