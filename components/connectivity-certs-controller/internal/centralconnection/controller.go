package centralconnection

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"time"

	"k8s.io/apimachinery/pkg/types"

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
	masterConnectionClient  ResourceClient
	certificatePreserver    certificates.Preserver
	certificateProvider     certificates.Provider
	mutualTLSClientProvider connectorservice.MutualTLSClientProvider
}

func InitCentralConnectionsController(
	mgr manager.Manager,
	appName string,
	certPreserver certificates.Preserver,
	certProvider certificates.Provider,
	mTLSClientProvider connectorservice.MutualTLSClientProvider) error {

	return startController(appName, mgr, certPreserver, certProvider, mTLSClientProvider)
}

func startController(
	appName string,
	mgr manager.Manager,
	certPreserver certificates.Preserver,
	certProvider certificates.Provider,
	mTLSClientProvider connectorservice.MutualTLSClientProvider) error {

	certRequestController := newCentralConnectionController(mgr.GetClient(), certPreserver, certProvider, mTLSClientProvider)

	c, err := controller.New(appName, mgr, controller.Options{Reconciler: certRequestController})
	if err != nil {
		return err
	}

	return c.Watch(&source.Kind{Type: &v1alpha1.CentralConnection{}}, &handler.EnqueueRequestForObject{})
}

func newCentralConnectionController(
	client ResourceClient,
	certPreserver certificates.Preserver,
	certProvider certificates.Provider,
	mTLSClientProvider connectorservice.MutualTLSClientProvider) *Controller {

	return &Controller{
		masterConnectionClient:  client,
		certificatePreserver:    certPreserver,
		certificateProvider:     certProvider,
		mutualTLSClientProvider: mTLSClientProvider,
	}
}

func (c *Controller) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	instance := &v1alpha1.CentralConnection{}

	log.Infof("Processing %s Central Connection", request.Name)

	err := c.masterConnectionClient.Get(context.Background(), request.NamespacedName, instance)
	if err != nil {
		return c.handleErrorWhileGettingInstance(err, request)
	}

	key, certificate, err := c.certificateProvider.GetClientCredentials()
	if err != nil {
		log.Errorf("Failed to read client certificate: %s", err.Error())
		return reconcile.Result{}, c.setErrorStatus(instance, err)
	}

	tlsConnectorClient := c.mutualTLSClientProvider.CreateClient(key, certificate)

	managementInfo, err := tlsConnectorClient.GetManagementInfo(instance.Spec.ManagementInfoURL)
	if err != nil {
		log.Errorf("Failed to request Management Info from URL %s: %s", instance.Spec.ManagementInfoURL, err.Error())
		return reconcile.Result{}, c.setErrorStatus(instance, err)
	}

	if c.shouldRenew() {
		err = c.renewCertificate(instance, tlsConnectorClient, managementInfo)
		if err != nil {
			log.Error(err)
			return reconcile.Result{}, c.setErrorStatus(instance, err)
		}
	}

	c.setSynchronizationStatus(instance)

	return reconcile.Result{}, c.updateCentralConnectionCR(instance)
}

func (c *Controller) handleErrorWhileGettingInstance(err error, request reconcile.Request) (reconcile.Result, error) {
	if k8sErrors.IsNotFound(err) {
		// TODO - should delete certs?
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

func (c *Controller) renewCertificate(connection *v1alpha1.CentralConnection, tlsConnectorClient connectorservice.MutualTLSConnectorClient, managementInfo connectorservice.ManagementInfo) error {
	renewedCerts, err := tlsConnectorClient.RenewCertificate(managementInfo.ManagementURLs.RenewalURL)
	if err != nil {
		return errors.Wrap(err, "Failed to renew client certificate")
	}

	err = c.certificatePreserver.PreserveCertificates(renewedCerts)
	if err != nil {
		return errors.Wrap(err, "Failed to save certificates to secrets")
	}

	pemBlock, _ := pem.Decode(renewedCerts.ClientCRT)
	if pemBlock == nil {
		return errors.New("Failed to decode client certificate pem")
	}

	clientCert, err := x509.ParseCertificate(pemBlock.Bytes)
	if err != nil {
		return errors.Wrap(err, "Failed to parse client certificate")
	}

	c.setCertificateStatus(connection, clientCert)

	return nil
}

func (c *Controller) setCertificateStatus(connection *v1alpha1.CentralConnection, certificate *x509.Certificate) {
	connection.Status.CertificateStatus = &v1alpha1.CertificateStatus{
		NotAfter:  metav1.NewTime(certificate.NotAfter),
		NotBefore: metav1.NewTime(certificate.NotBefore),
	}
}

func (c *Controller) setSynchronizationStatus(connection *v1alpha1.CentralConnection) {
	syncTime := metav1.Now()

	connection.Status.SynchronizationStatus = v1alpha1.SynchronizationStatus{
		LastSync:    syncTime,
		LastSuccess: syncTime,
	}

	connection.Status.Error = nil
}

func (c *Controller) setErrorStatus(connection *v1alpha1.CentralConnection, err error) error {
	syncTime := metav1.NewTime(time.Now())

	connection.Status.Error = &v1alpha1.CentralConnectionError{Message: err.Error()}
	connection.Status.SynchronizationStatus.LastSync = syncTime
	connection.Status.CertificateStatus = nil

	updateError := c.updateCentralConnectionCR(connection)
	if updateError != nil {
		log.Errorf("Failed to set error status on %s Connection", connection.Name)
		return updateError
	}

	return err
}

// TODO - fix retryOnConflict - consider proper client
func (c *Controller) updateCentralConnectionCR(connection *v1alpha1.CentralConnection) error {
	name := types.NamespacedName{Name: connection.Name, Namespace: connection.Namespace}

	existingConnection := &v1alpha1.CentralConnection{}

	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		err := c.masterConnectionClient.Get(context.Background(), name, existingConnection)
		if err != nil {
			return errors.Wrap(err, "Failed to get up to date resource")
		}

		existingConnection.Status = connection.Status

		return c.masterConnectionClient.Update(context.Background(), existingConnection)
	})
	if err != nil {
		return errors.Wrap(err, "Failed to update Central Connection")
	}

	return nil
}
