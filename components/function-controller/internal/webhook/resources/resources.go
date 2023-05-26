package resources

import (
	"context"
	"os"
	"path"
	"time"

	"go.uber.org/zap"

	"github.com/pkg/errors"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

func SetupResourcesController(ctx context.Context, mgr ctrl.Manager, serviceName, serviceNamespace, secretName string, log *zap.SugaredLogger) error {
	logger := log.Named("resource-ctrl")
	certPath := path.Join(DefaultCertDir, CertFile)
	certBytes, err := os.ReadFile(certPath)
	if err != nil {
		return errors.Wrapf(err, "failed to read caBundle file: %s", certPath)
	}

	webhookConfig := WebhookConfig{
		CABundle:         certBytes,
		ServiceName:      serviceName,
		ServiceNamespace: serviceNamespace,
	}
	// We are going to talk to the API server _before_ we start the manager.
	// Since the default manager client reads from cache, we will get an error.
	// So, we create a "serverClient" that would read from the API directly.
	// We only use it here, this only runs at start up, so it shouldn't be to much for the API
	serverClient, err := ctrlclient.New(ctrl.GetConfigOrDie(), ctrlclient.Options{})
	if err != nil {
		return errors.Wrap(err, "failed to create a server client")
	}

	logger.Info("initializing the defaulting webhook configuration")
	if err := InjectCABundleIntoWebhooks(ctx, serverClient, webhookConfig, MutatingWebhook); err != nil {
		return errors.Wrap(err, "failed to ensure defaulting webhook configuration")
	}

	logger.Info("initializing the validation webhook configuration")
	if err := InjectCABundleIntoWebhooks(ctx, serverClient, webhookConfig, ValidatingWebHook); err != nil {
		return errors.Wrap(err, "failed to ensure validating webhook configuration")
	}
	// watch over the configuration
	logger.Info("creating webhook resources controller")
	c, err := controller.New("webhook-resources-controller", mgr, controller.Options{
		Reconciler: &resourceReconciler{
			webhookConfig: webhookConfig,
			client:        mgr.GetClient(),
			secretName:    secretName,
			logger:        log.Named("webhook-resource-controller"),
		},
	})
	if err != nil {
		return errors.Wrap(err, "failed to create webhook-config-controller")
	}

	if err := c.Watch(&source.Kind{
		Type: &admissionregistrationv1.ValidatingWebhookConfiguration{}},
		&handler.EnqueueRequestForObject{},
	); err != nil {
		return errors.Wrap(err, "failed to watch ValidatingWebhookConfiguration")
	}

	if err := c.Watch(&source.Kind{
		Type: &admissionregistrationv1.MutatingWebhookConfiguration{}},
		&handler.EnqueueRequestForObject{},
	); err != nil {
		return errors.Wrap(err, "failed to watch MutatingWebhookConfiguration")
	}
	if err := c.Watch(&source.Kind{
		Type: &corev1.Secret{}},
		&handler.EnqueueRequestForObject{},
	); err != nil {
		return errors.Wrap(err, "failed to watch Secrets")
	}
	return nil
}

type resourceReconciler struct {
	webhookConfig WebhookConfig
	secretName    string
	client        ctrlclient.Client
	logger        *zap.SugaredLogger
}

func (r *resourceReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	// if the request is not one of our managed resources, we bail.
	secretNamespaced := types.NamespacedName{Name: r.secretName, Namespace: r.webhookConfig.ServiceNamespace}
	if request.Name != DefaultingWebhookName &&
		request.Name != ValidationWebhookName &&
		request.NamespacedName.String() != secretNamespaced.String() {
		return reconcile.Result{}, nil
	}

	r.logger.With("name", request.Name).Info("reconciling webhook resources")
	if err := r.reconcilerWebhooks(ctx, request); err != nil {
		return reconcile.Result{}, errors.Wrap(err, "failed to reconcile webhook resources")
	}
	result, err := r.reconcilerSecret(ctx, request)
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "failed to reconcile webhook resources")
	}
	if result == Updated {
		r.logger.Info("certificate updated successfully, restarting")
		//This is not an elegant solution, but the webhook need to reconfigure itself to use updated certificate.
		//Cert-watcher from controller-runtime should refresh the certificate, but it doesn't work.
		os.Exit(0)
	}
	r.logger.With("name", request.Name).Info("webhook resources reconciled successfully")
	return reconcile.Result{RequeueAfter: 1 * time.Hour}, nil
}

func (r *resourceReconciler) reconcilerWebhooks(ctx context.Context, request reconcile.Request) error {
	if request.Name == DefaultingWebhookName {
		r.logger.Info("reconciling webhook defaulting webhook configuration")
		if err := InjectCABundleIntoWebhooks(ctx, r.client, r.webhookConfig, MutatingWebhook); err != nil {
			return errors.Wrap(err, "failed to ensure defaulting webhook configuration")
		}
	}
	if request.Name == ValidationWebhookName {
		r.logger.Info("reconciling webhook validating webhook configuration")
		if err := InjectCABundleIntoWebhooks(ctx, r.client, r.webhookConfig, ValidatingWebHook); err != nil {
			return errors.Wrap(err, "failed to ensure validating webhook configuration")
		}
	}
	return nil
}

func (r *resourceReconciler) reconcilerSecret(ctx context.Context, request reconcile.Request) (Result, error) {
	ctrl.LoggerFrom(ctx).Info("reconciling webhook secret")
	secretNamespaced := types.NamespacedName{Name: r.secretName, Namespace: r.webhookConfig.ServiceNamespace}
	if request.NamespacedName.String() != secretNamespaced.String() {
		return NoResult, nil
	}
	result, err := EnsureWebhookSecret(ctx, r.client, request.Name, request.Namespace, r.webhookConfig.ServiceName, r.logger)
	if err != nil {
		return NoResult, errors.Wrap(err, "failed to reconcile webhook secret")
	}
	return result, nil
}
