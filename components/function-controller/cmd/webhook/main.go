package main

import (
	"context"
	"io/ioutil"
	"path"

	"github.com/go-logr/zapr"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"

	"github.com/kyma-project/kyma/components/function-controller/internal/webhook"
	"github.com/kyma-project/kyma/components/function-controller/internal/webhook/resources"
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	ctrlzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	ctrlwebhook "sigs.k8s.io/controller-runtime/pkg/webhook"
)

type config struct {
	SystemNamespace    string `envconfig:"default=kyma-system"`
	WebhookServiceName string `envconfig:"default=serverless-webhook"`
	WebhookSecretName  string `envconfig:"default=serverless-webhook"`
	WebhookPort        int    `envconfig:"default=8443"`
}

const (
	caBundleFile = "ca-cert.pem"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = serverlessv1alpha1.AddToScheme(scheme)
	_ = admissionregistrationv1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme

	setupControllerLogger()
}

func main() {
	cfg := &config{}
	if err := envconfig.Init(cfg); err != nil {
		panic(errors.Wrap(err, "while reading env variables"))
	}
	// manager setup
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		Port:               cfg.WebhookPort,
		MetricsBindAddress: ":9090",
	})
	if err != nil {
		panic(err)
	}
	// webhook server setup
	whs := mgr.GetWebhookServer()
	whs.CertName = "server-cert.pem"
	whs.KeyName = "server-key.pem"
	whs.Register("/defaulting",
		&ctrlwebhook.Admission{
			Handler: webhook.NewDefaultingWebhook(webhook.ReadDefaultingConfig(), mgr.GetClient()),
		},
	)
	whs.Register("/validation",
		&ctrlwebhook.Admission{
			Handler: webhook.NewValidatingHook(webhook.ReadValidationConfig(), mgr.GetClient()),
		},
	)
	// apply and monitor configuration
	if err := setupWebhookConfigurations(context.Background(), mgr, *cfg); err != nil {
		panic(err)
	}
	// start the server manager:q
	err = mgr.Start(ctrl.SetupSignalHandler())
	if err != nil {
		panic(err)
	}
}

func setupControllerLogger() {
	atomicLevel := zap.NewAtomicLevelAt(zapcore.InfoLevel)
	zapLogger := ctrlzap.NewRaw(ctrlzap.UseDevMode(true), ctrlzap.Level(&atomicLevel))
	ctrl.SetLogger(zapr.NewLogger(zapLogger))
}

func setupWebhookConfigurations(ctx context.Context, mgr ctrl.Manager, cfg config) error {
	caPath := path.Join(mgr.GetWebhookServer().CertDir, caBundleFile)
	caBundle, err := ioutil.ReadFile(caPath)
	if err != nil {
		return errors.Wrapf(err, "failed to read caBundel file: %s", caBundle)
	}

	webhookConfig := resources.WebhookConfig{
		Type:             resources.MutatingWebhook,
		CABundel:         caBundle,
		ServiceName:      cfg.WebhookServiceName,
		ServiceNamespace: cfg.SystemNamespace,
		Port:             int32(cfg.WebhookPort),
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
	if err := resources.EnsureWebhookConfigurationFor(ctx, serverClient, webhookConfig, resources.MutatingWebhook); err != nil {
		return errors.Wrap(err, "failed to ensure defaulting webhook configuration")
	}
	if err := resources.EnsureWebhookConfigurationFor(ctx, serverClient, webhookConfig, resources.ValidatingWebHook); err != nil {
		return errors.Wrap(err, "failed to ensure validating webhook configuration")
	}
	// watch over the configuration
	c, err := controller.New("webhook-config-controller", mgr, controller.Options{
		Reconciler: NewWebhookConfig(webhookConfig, mgr.GetClient()),
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

type WebhookConfig struct {
	config resources.WebhookConfig
	client ctrlclient.Client
}

func NewWebhookConfig(config resources.WebhookConfig, client ctrlclient.Client) *WebhookConfig {
	return &WebhookConfig{
		config: config,
		client: client,
	}
}

func (r *WebhookConfig) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	if request.Name != resources.DefaultingWebhookName &&
		request.Name != resources.ValidationWebhookName {
		return reconcile.Result{}, nil

	}

	if request.Name == resources.DefaultingWebhookName {
		return reconcile.Result{}, resources.EnsureWebhookConfigurationFor(ctx, r.client, r.config, resources.MutatingWebhook)
	}
	if request.Name == resources.ValidationWebhookName {
		return reconcile.Result{}, resources.EnsureWebhookConfigurationFor(ctx, r.client, r.config, resources.ValidatingWebHook)
	}

	return reconcile.Result{}, nil
}
