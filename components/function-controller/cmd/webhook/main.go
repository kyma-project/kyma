package main

import (
	"context"

	"github.com/kyma-project/kyma/components/function-controller/internal/webhook"
	"github.com/kyma-project/kyma/components/function-controller/internal/webhook/resources"
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"

	"github.com/go-logr/zapr"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
	ctrlwebhook "sigs.k8s.io/controller-runtime/pkg/webhook"
)

type config struct {
	SystemNamespace    string `envconfig:"default=kyma-system"`
	WebhookServiceName string `envconfig:"default=serverless-webhook"`
	WebhookSecretName  string `envconfig:"default=serverless-webhook"`
	WebhookPort        int    `envconfig:"default=8443"`
}

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = serverlessv1alpha1.AddToScheme(scheme)
	_ = admissionregistrationv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
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

	if err := resources.SetupCertificates(
		context.Background(),
		cfg.WebhookSecretName,
		cfg.SystemNamespace,
		cfg.WebhookServiceName); err != nil {
		panic(err)
	}
	// webhook server setup
	whs := mgr.GetWebhookServer()
	whs.CertName = resources.CertFile
	whs.KeyName = resources.KeyfFile
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
	if err := resources.SetupResourcesController(
		context.Background(),
		mgr,
		cfg.WebhookServiceName,
		cfg.SystemNamespace,
		cfg.WebhookSecretName,
	); err != nil {
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
