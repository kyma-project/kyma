package resources

import (
	"context"
	"io/ioutil"
	"path"

	"github.com/pkg/errors"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

func SetupResourcesController(ctx context.Context, mgr ctrl.Manager, serviceName, serviceNamespace, secretName string) error {
	certPath := path.Join(DefaultCertDir, CertFile)
	certBytes, err := ioutil.ReadFile(certPath)
	if err != nil {
		return errors.Wrapf(err, "failed to read caBundel file: %s", certBytes)
	}

	webhookConfig := WebhookConfig{
		Type:             MutatingWebhook,
		CABundel:         certBytes,
		ServiceName:      serviceName,
		ServiceNamespace: serviceNamespace,
	}
	// initial webhook configuration

	// We are going to talk to the API server _before_ we start the manager.
	// Since the default manager client reads from cache, we will get an error.
	// So, we create a "serverClient" that would read from the API directly.
	// We only use it here, this only runs at start up, so it shouldn't bee to much for the API
	serverClient, err := ctrlclient.New(ctrl.GetConfigOrDie(), ctrlclient.Options{})
	if err != nil {
		return errors.Wrap(err, "failed to create a server client")
	}
	if err := EnsureWebhookConfigurationFor(ctx, serverClient, webhookConfig, MutatingWebhook); err != nil {
		return errors.Wrap(err, "failed to ensure defaulting webhook configuration")
	}
	if err := EnsureWebhookConfigurationFor(ctx, serverClient, webhookConfig, ValidatingWebHook); err != nil {
		return errors.Wrap(err, "failed to ensure validating webhook configuration")
	}
	// watch over the configuration
	c, err := controller.New("webhook-config-controller", mgr, controller.Options{
		Reconciler: &resourceReconciler{
			webhookConfig: webhookConfig,
			client:        mgr.GetClient(),
			secretName:    secretName,
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
	return nil
}

type resourceReconciler struct {
	webhookConfig WebhookConfig
	secretName    string
	client        ctrlclient.Client
}

func (r *resourceReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	// if the request is not one of our managed resources, we bail.
	secretNamespaced := types.NamespacedName{Name: r.secretName, Namespace: r.webhookConfig.ServiceNamespace}
	if request.Name != DefaultingWebhookName &&
		request.Name != ValidationWebhookName &&
		request.NamespacedName.String() != secretNamespaced.String() {
		return reconcile.Result{}, nil

	}
	if err := r.reconcilerWebhooks(ctx, request); err != nil {
		return reconcile.Result{}, err
	}
	if err := r.reconcilerSecret(ctx, request); err != nil {
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

func (r *resourceReconciler) reconcilerWebhooks(ctx context.Context, request reconcile.Request) error {
	if request.Name == DefaultingWebhookName {
		if err := EnsureWebhookConfigurationFor(ctx, r.client, r.webhookConfig, MutatingWebhook); err != nil {
			return err
		}
	}
	if request.Name == ValidationWebhookName {
		if err := EnsureWebhookConfigurationFor(ctx, r.client, r.webhookConfig, ValidatingWebHook); err != nil {
			return err
		}
	}
	return nil
}

func (r *resourceReconciler) reconcilerSecret(ctx context.Context, request reconcile.Request) error {
	secretNamespaced := types.NamespacedName{Name: r.secretName, Namespace: r.webhookConfig.ServiceNamespace}
	if request.NamespacedName.String() == secretNamespaced.String() {
		return EnsureWebhookSecret(ctx, r.client, request.Name, request.Namespace, r.webhookConfig.ServiceName)
	}
	return nil
}
