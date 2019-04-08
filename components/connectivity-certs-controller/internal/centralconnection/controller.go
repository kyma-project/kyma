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
	mutualTLSClientProvider connectorservice.MutualTLSClientProvider
	minimalSyncTime         time.Duration
	logger                  *log.Entry
}

func InitCentralConnectionsController(
	mgr manager.Manager,
	appName string,
	minimalSyncTime time.Duration,
	certPreserver certificates.Preserver,
	mTLSClientProvider connectorservice.MutualTLSClientProvider) error {

	return startController(appName, mgr, minimalSyncTime, certPreserver, mTLSClientProvider)
}

func startController(
	appName string,
	mgr manager.Manager,
	minimalSyncTime time.Duration,
	certPreserver certificates.Preserver,
	mTLSClientProvider connectorservice.MutualTLSClientProvider) error {

	certRequestController := newCentralConnectionController(mgr.GetClient(), minimalSyncTime, certPreserver, mTLSClientProvider)

	c, err := controller.New(appName, mgr, controller.Options{Reconciler: certRequestController})
	if err != nil {
		return err
	}

	return c.Watch(&source.Kind{Type: &v1alpha1.CentralConnection{}}, &handler.EnqueueRequestForObject{})
}

func newCentralConnectionController(
	client ResourceClient,
	minimalSyncTime time.Duration,
	certPreserver certificates.Preserver,
	mTLSClientProvider connectorservice.MutualTLSClientProvider) *Controller {

	return &Controller{
		masterConnectionClient:  client,
		minimalSyncTime:         minimalSyncTime,
		certificatePreserver:    certPreserver,
		mutualTLSClientProvider: mTLSClientProvider,
		logger:                  log.WithField("Controller", "Central Connection"),
	}
}

func (c *Controller) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	instance := &v1alpha1.CentralConnection{}

	c.logger.Infof("Processing %s Central Connection", request.Name)

	err := c.masterConnectionClient.Get(context.Background(), request.NamespacedName, instance)
	if err != nil {
		return c.handleErrorWhileGettingInstance(err, request)
	}

	if !c.shouldSynchronizeConnection(instance) {
		c.logger.Infof("Skipping synchronization of %s Central Connection. Last sync: %v", instance.Name, instance.Status.SynchronizationStatus.LastSync)
		return reconcile.Result{}, nil
	}

	tlsConnectorClient, err := c.mutualTLSClientProvider.CreateClient()
	if err != nil {
		c.logger.Errorf("Failed to create mutual TLS Connector Client: %s", err.Error())
		return reconcile.Result{}, c.setErrorStatus(instance, err)
	}

	c.logger.Infof("Trying to access Management Info for %s Central Connection...", instance.Name)
	managementInfo, err := tlsConnectorClient.GetManagementInfo(instance.Spec.ManagementInfoURL)
	if err != nil {
		c.logger.Errorf("Failed to request Management Info from URL %s: %s", instance.Spec.ManagementInfoURL, err.Error())
		return reconcile.Result{}, c.setErrorStatus(instance, err)
	}
	c.logger.Infof("Successfully fetched Management Info for %s Central Connection", instance.Name)

	if c.shouldRenew(instance) {
		c.logger.Infof("Trying to renew client certificate for %s Central Connection...", instance.Name)
		err = c.renewCertificate(instance, tlsConnectorClient, managementInfo)
		if err != nil {
			c.logger.Error(err)
			return reconcile.Result{}, c.setErrorStatus(instance, err)
		}
		c.logger.Infof("Successfully renewed Certificate for %s Central Connection", instance.Name)
	}

	c.setSynchronizationStatus(instance)

	return reconcile.Result{}, c.updateCentralConnectionCR(instance)
}

func (c *Controller) handleErrorWhileGettingInstance(err error, request reconcile.Request) (reconcile.Result, error) {
	if k8sErrors.IsNotFound(err) {
		c.logger.Infof("Connection %s has been deleted", request.Name)
		return reconcile.Result{}, nil
	}
	c.logger.Errorf("Error while getting instance: %s", err.Error())
	return reconcile.Result{}, err
}

func (c *Controller) shouldSynchronizeConnection(connection *v1alpha1.CentralConnection) bool {
	timeFromLastSync := time.Since(connection.Status.SynchronizationStatus.LastSync.Time)
	return timeFromLastSync > c.minimalSyncTime
}

func (c *Controller) shouldRenew(connection *v1alpha1.CentralConnection) bool {
	//return connection.Status.CertificateStatus.NotAfter.Time.Second() < (time.Now() + c.minimalSyncTime)
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
		c.logger.Errorf("Failed to set error status on %s Connection", connection.Name)
		return updateError
	}

	return err
}

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
