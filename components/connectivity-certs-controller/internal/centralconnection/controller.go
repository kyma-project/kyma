package centralconnection

import (
	"context"
	"crypto/x509"
	"time"

	"github.com/pkg/errors"

	"k8s.io/client-go/util/retry"

	"github.com/kyma-project/kyma/components/connectivity-certs-controller/internal/certificates"
	"github.com/kyma-project/kyma/components/connectivity-certs-controller/internal/connectorservice"
	"github.com/kyma-project/kyma/components/connectivity-certs-controller/pkg/apis/applicationconnector/v1alpha1"

	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
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
	certificateProvider    certificates.CertificateProvider
	csrProvider            certificates.CSRProvider
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
	instance := &v1alpha1.CentralConnection{}

	log.Infof("Processing %s Central Connection", request.Name)

	err := c.masterConnectionClient.Get(context.Background(), request.NamespacedName, instance)
	if err != nil {
		return c.handleErrorWhileGettingInstance(err, request)
	}

	if instance.Status.Error != nil {
		// TODO - what to do in case of error?
		//log.Infof("Certificate Request %s has an error status, certificate will not be fetched", instance.Name)
		return reconcile.Result{}, nil
	}

	key, certificate, err := c.certificateProvider.GetClientCredentials()
	if err != nil {
		log.Errorf("Failed to read client certificate: %s", err.Error())
		return reconcile.Result{}, c.setErrorStatus(instance, err)
	}

	tlsConnectorClient := connectorservice.NewMutualTLSConnectorClient(key, []*x509.Certificate{certificate})

	managementInfo, err := tlsConnectorClient.GetManagementInfo(instance.Spec.ManagementInfoURL)
	if err != nil {
		log.Errorf("Failed to request Management Info from URL %s: %s", instance.Spec.ManagementInfoURL, err.Error())
		return reconcile.Result{}, c.setErrorStatus(instance, err)
	}

	if c.shouldRenew() {
		err = c.renewCertificate(tlsConnectorClient, managementInfo, certificate)
		if err != nil {
			log.Error(err)
			return reconcile.Result{}, c.setErrorStatus(instance, err)
		}

		c.setRenewalStatus(instance, certificate)
	}

	c.setSynchronizationStatus(instance)

	return reconcile.Result{}, c.updateCentralConnectionCR(instance)
}

func (c *Controller) handleErrorWhileGettingInstance(err error, request reconcile.Request) (reconcile.Result, error) {
	if k8sErrors.IsNotFound(err) {
		log.Infof("Connection %s has been deleted", request.Name)
		return reconcile.Result{}, nil
	}
	log.Errorf("Error while getting instance: %s", err.Error())
	return reconcile.Result{}, err
}

func (c *Controller) shouldRenew() bool {
	// TODO - renew only if needed
	// TODO - how do we decide if it is time to renew?plan
	return true
}

func (c *Controller) renewCertificate(tlsConnectorClient connectorservice.MutualTLSConnectorClient, managementInfo connectorservice.ManagementInfo, certificate *x509.Certificate) error {
	// TODO: Subject should be returned from Management Info in the future
	subject := certificate.Subject

	csr, err := c.csrProvider.CreateCSR(subject)
	if err != nil {
		return errors.Wrap(err, "Failed to create CSR")
	}

	renewedCerts, err := tlsConnectorClient.RenewCertificate(managementInfo.ManagementURLs.RenewalURL, csr)
	if err != nil {
		return errors.Wrap(err, "Failed to renew client certificate")
	}

	err = c.certificatePreserver.PreserveCertificates(renewedCerts)
	if err != nil {
		return errors.Wrap(err, "Failed to save certificates to secrets")
	}

	return nil
}

func (c *Controller) setRenewalStatus(connection *v1alpha1.CentralConnection, certificate *x509.Certificate) {
	connection.Status.CertificateStatus = v1alpha1.CertificateStatus{
		ValidTo:  metav1.NewTime(certificate.NotAfter),
		IssuedAt: metav1.NewTime(time.Now()),
	}
}

func (c *Controller) setSynchronizationStatus(connection *v1alpha1.CentralConnection) {
	syncTime := metav1.NewTime(time.Now())

	connection.Status.SynchronizationStatus = v1alpha1.SynchronizationStatus{
		LastSync:    syncTime,
		LastSuccess: syncTime,
	}

	connection.Status.Error = nil
}

func (c *Controller) setErrorStatus(connection *v1alpha1.CentralConnection, err error) error {
	syncTime := metav1.NewTime(time.Now())

	connection.Status.Error.Message = err.Error()
	connection.Status.SynchronizationStatus.LastSync = syncTime
	// TODO - remove cert status?

	return c.updateCentralConnectionCR(connection)
}

func (r *Controller) updateCentralConnectionCR(connection *v1alpha1.CentralConnection) error {
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		return r.masterConnectionClient.Update(context.Background(), connection)
	})

	if err != nil {
		return errors.Wrap(err, "Failed to update Central Connection")
	}

	return nil
}
