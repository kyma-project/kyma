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

const (
	certValidityRenewalThreshold = 0.3
)

type ResourceClient interface {
	Get(ctx context.Context, key client.ObjectKey, obj runtime.Object) error
	Update(ctx context.Context, obj runtime.Object) error
}

type Controller struct {
	centralConnectionClient             ResourceClient
	certificatePreserver                certificates.Preserver
	certificateProvider                 certificates.Provider
	establishedConnectionClientProvider connectorservice.EstablishedConnectionClientProvider
	minimalConnectionSyncPeriod         time.Duration
	logger                              *log.Entry
}

func InitCentralConnectionController(
	mgr manager.Manager,
	appName string,
	minimalConnectionSyncPeriod time.Duration,
	certPreserver certificates.Preserver,
	certProvider certificates.Provider,
	connectionClientProvider connectorservice.EstablishedConnectionClientProvider) error {

	return startController(appName, mgr, minimalConnectionSyncPeriod, certPreserver, certProvider, connectionClientProvider)
}

func startController(
	appName string,
	mgr manager.Manager,
	minimalConnectionSyncPeriod time.Duration,
	certPreserver certificates.Preserver,
	certProvider certificates.Provider,
	connectionClientProvider connectorservice.EstablishedConnectionClientProvider) error {

	certRequestController := newCentralConnectionController(mgr.GetClient(), minimalConnectionSyncPeriod, certPreserver, certProvider, connectionClientProvider)

	c, err := controller.New(appName, mgr, controller.Options{Reconciler: certRequestController})
	if err != nil {
		return err
	}

	return c.Watch(&source.Kind{Type: &v1alpha1.CentralConnection{}}, &handler.EnqueueRequestForObject{})
}

func newCentralConnectionController(
	client ResourceClient,
	minimalConnectionSyncPeriod time.Duration,
	certPreserver certificates.Preserver,
	certProvider certificates.Provider,
	connectionClientProvider connectorservice.EstablishedConnectionClientProvider) *Controller {

	return &Controller{
		centralConnectionClient:             client,
		minimalConnectionSyncPeriod:         minimalConnectionSyncPeriod,
		certificatePreserver:                certPreserver,
		certificateProvider:                 certProvider,
		establishedConnectionClientProvider: connectionClientProvider,
		logger: log.WithField("Controller", "Central Connection"),
	}
}

func (c *Controller) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	instance := &v1alpha1.CentralConnection{}

	c.logger.Infof("Processing %s Central Connection", request.Name)

	err := c.centralConnectionClient.Get(context.Background(), request.NamespacedName, instance)
	if err != nil {
		return c.handleErrorWhileGettingInstance(err, request)
	}

	if !c.shouldSynchronizeConnection(instance) {
		c.logger.Infof("Skipping synchronization of %s Central Connection. Last sync: %v", instance.Name, instance.Status.SynchronizationStatus.LastSync)
		return reconcile.Result{}, nil
	}

	certCredentials, err := c.getCertificateCredentials()
	if err != nil {
		log.Infof("Failed to get certificates for %s Central Connection: %s", instance.Name, err.Error())
		return reconcile.Result{}, c.setErrorStatus(instance, err)
	}

	if !instance.HasCertStatus() {
		c.setCertificateStatus(instance, certCredentials.ClientCert)
	}

	err = c.synchronizeWithConnector(instance, certCredentials)
	if err != nil {
		log.Infof("Failed to synchronize %s Central Connection with Connector Service: %s", instance.Name, err.Error())
		return reconcile.Result{}, c.setErrorStatus(instance, err)
	}

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
	if connection.Spec.RenewNow {
		return true
	}

	timeFromLastSync := time.Since(connection.Status.SynchronizationStatus.LastSync.Time)
	return timeFromLastSync > c.minimalConnectionSyncPeriod
}

func (c *Controller) synchronizeWithConnector(connection *v1alpha1.CentralConnection, certCredentials connectorservice.CertificateCredentials) error {
	establishedConnectionClient := c.establishedConnectionClientProvider.CreateClient(certCredentials)

	c.logger.Infof("Trying to access Management Info for %s Central Connection...", connection.Name)
	managementInfo, err := establishedConnectionClient.GetManagementInfo(connection.Spec.ManagementInfoURL)
	if err != nil {
		c.logger.Errorf("Failed to request Management Info from URL %s for %s Central Connection: %s", connection.Spec.ManagementInfoURL, connection.Name, err.Error())
		return errors.Wrap(err, "Failed to access Management Info")
	}
	c.logger.Infof("Successfully fetched Management Info for %s Central Connection", connection.Name)

	c.setSynchronizationStatus(connection)

	if !shouldRenew(connection, c.minimalConnectionSyncPeriod) {
		c.logger.Infof("Skipping certificate renewal for %s Central Connection", connection.Name)
		return nil
	}

	c.logger.Infof("Trying to renew client certificate for %s Central Connection...", connection.Name)
	err = c.renewCertificate(connection, establishedConnectionClient, managementInfo)
	if err != nil {
		c.logger.Errorf("Failed to renew certificate for %s Central Connection: %s", connection.Name, err.Error())
		return errors.Wrap(err, "Failed to renew certificate")
	}
	c.logger.Infof("Successfully renewed Certificate for %s Central Connection", connection.Name)

	return nil
}

// Certificate should be renewed when less than 30% of validity time is left or if time left is less than 2 times minimalConnectionSyncPeriod
func shouldRenew(connection *v1alpha1.CentralConnection, minimalSyncTime time.Duration) bool {
	if connection.Spec.RenewNow {
		return true
	}

	notBefore := connection.Status.CertificateStatus.NotBefore.Unix()
	notAfter := connection.Status.CertificateStatus.NotAfter.Unix()

	certValidity := notAfter - notBefore

	timeLeft := float64(notAfter - time.Now().Unix())

	return timeLeft < float64(certValidity)*certValidityRenewalThreshold || timeLeft < 2*minimalSyncTime.Seconds()
}

func (c *Controller) getCertificateCredentials() (connectorservice.CertificateCredentials, error) {
	clientKey, clientCert, err := c.certificateProvider.GetClientCredentials()
	if err != nil {
		return connectorservice.CertificateCredentials{}, errors.Wrap(err, "Failed to get client key and certificate")
	}

	caCertificates, err := c.certificateProvider.GetCACertificates()
	if err != nil {
		return connectorservice.CertificateCredentials{}, errors.Wrap(err, "Failed to get CA certificate")
	}

	return connectorservice.CertificateCredentials{
		ClientKey:  clientKey,
		ClientCert: clientCert,
		CACerts:    caCertificates,
	}, nil
}

func (c *Controller) renewCertificate(connection *v1alpha1.CentralConnection, establishedConnectionClient connectorservice.EstablishedConnectionClient, managementInfo connectorservice.ManagementInfo) error {
	renewedCerts, err := establishedConnectionClient.RenewCertificate(managementInfo.ManagementURLs.RenewalURL)
	if err != nil {
		return errors.Wrap(err, "Failed to renew client certificate")
	}

	err = c.certificatePreserver.PreserveCertificates(renewedCerts)
	if err != nil {
		return errors.Wrap(err, "Failed to save certificates to secrets")
	}

	clientCert, err := parseCertificate(renewedCerts.ClientCRT)
	if err != nil {
		return errors.Wrap(err, "Failed to parse client certificate")
	}

	c.setCertificateStatus(connection, clientCert)
	connection.Spec.RenewNow = false

	return nil
}

func parseCertificate(rawPem []byte) (*x509.Certificate, error) {
	pemBlock, _ := pem.Decode(rawPem)
	if pemBlock == nil {
		return nil, errors.New("Failed to decode client certificate pem")
	}

	return x509.ParseCertificate(pemBlock.Bytes)
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

	connection.Spec.RenewNow = false

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
		err := c.centralConnectionClient.Get(context.Background(), name, existingConnection)
		if err != nil {
			return errors.Wrap(err, "Failed to get up to date resource")
		}

		existingConnection.Status = connection.Status
		existingConnection.Spec.RenewNow = connection.Spec.RenewNow

		return c.centralConnectionClient.Update(context.Background(), existingConnection)
	})
	if err != nil {
		return errors.Wrap(err, "Failed to update Central Connection")
	}

	return nil
}
