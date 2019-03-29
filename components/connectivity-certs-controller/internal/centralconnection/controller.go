package masterconnection

import (
	"context"

	"github.com/kyma-project/kyma/components/connectivity-certs-controller/internal/certificates"
	"github.com/kyma-project/kyma/components/connectivity-certs-controller/internal/connectorservice"
	"github.com/kyma-project/kyma/components/connectivity-certs-controller/pkg/apis/applicationconnector/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type ResourceClient interface {
	Get(ctx context.Context, key client.ObjectKey, obj runtime.Object) error
	Update(ctx context.Context, obj runtime.Object) error
	Delete(ctx context.Context, obj runtime.Object, opts ...client.DeleteOptionFunc) error
}

type Controller struct {
	masterConnectionClient ResourceClient
	connectorClient        connectorservice.Client
	certificatePreserver   certificates.Preserver
}

func InitMasterConnectionsController(mgr manager.Manager, appName string, connectorClient connectorservice.Client, certPreserver certificates.Preserver) error {
	return startController(appName, mgr, connectorClient, certPreserver)
}

func startController(appName string, mgr manager.Manager, connectorClient connectorservice.Client, certPreserver certificates.Preserver) error {
	certRequestController := newMasterConnectionController(mgr.GetClient(), connectorClient, certPreserver)

	c, err := controller.New(appName, mgr, controller.Options{Reconciler: certRequestController})
	if err != nil {
		return err
	}

	return c.Watch(&source.Kind{Type: &v1alpha1.CentralConnection{}}, &handler.EnqueueRequestForObject{})
}

func newMasterConnectionController(client ResourceClient, connectorClient connectorservice.Client, certPreserver certificates.Preserver) *Controller {
	return &Controller{
		masterConnectionClient: client,
		connectorClient:        connectorClient,
		certificatePreserver:   certPreserver,
	}
}

func (c *Controller) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	//instance := &v1alpha1.CertificateRequest{}
	//
	//log.Infof("Processing %s request", request.Name)
	//
	//err := c.masterConnectionClient.Get(context.Background(), request.NamespacedName, instance)
	//if err != nil {
	//	return c.handleErrorWhileGettingInstance(err, request)
	//}
	//
	//if instance.Status.Error != "" {
	//	log.Infof("Certificate Request %s has an error status, certificate will not be fetched", instance.Name)
	//	return reconcile.Result{}, nil
	//}
	//
	//certs, err := c.connectorClient.RequestCertificates(instance.Spec.CSRInfoURL)
	//if err != nil {
	//	log.Errorf("Error while requesting certificates from Connector Service: %s", err.Error())
	//	return reconcile.Result{}, c.setRequestErrorStatus(instance, err)
	//}
	//log.Infoln("Certificates fetched successfully")
	//
	//err = c.certificatePreserver.PreserveCertificates(certs)
	//if err != nil {
	//	log.Errorf("Error while saving certificates to secrets: %s", err.Error())
	//	return reconcile.Result{}, c.setRequestErrorStatus(instance, err)
	//}
	//log.Infoln("Certificates saved successfully")
	//
	//err = c.masterConnectionClient.Delete(context.Background(), instance)
	//if err != nil {
	//	return reconcile.Result{}, err
	//}
	//log.Infof("CertificatesRequest %s successfully deleted", instance.Name)

	return reconcile.Result{}, nil
}

func (c *Controller) handleErrorWhileGettingInstance(err error, request reconcile.Request) (reconcile.Result, error) {
	//if k8sErrors.IsNotFound(err) {
	//	log.Infof("Request %s has been deleted", request.Name)
	//	return reconcile.Result{}, nil
	//}
	//log.Errorf("Error while getting instance, %s", err.Error())
	return reconcile.Result{}, err
}

func (c *Controller) setRequestErrorStatus(instance *v1alpha1.CertificateRequest, statusError error) error {
	//instance.Status = v1alpha1.CertificateRequestStatus{
	//	Error: statusError.Error(),
	//}
	//
	//err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
	//	return c.masterConnectionClient.Update(context.Background(), instance)
	//})
	//if err != nil {
	//	return err
	//}

	return nil
}
