package main

import (
	"context"
	"os"

	"github.com/kyma-project/kyma/components/function-controller/internal/webhook"
	"github.com/kyma-project/kyma/components/function-controller/internal/webhook/resources"
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"

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

//nolint
func init() {
	_ = serverlessv1alpha2.AddToScheme(scheme)
	_ = serverlessv1alpha1.AddToScheme(scheme)
	_ = admissionregistrationv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme

	setupControllerLogger()
}

func main() {
	log := ctrl.Log.WithName("setup:")

	cfg := &config{}
	log.Info("reading configuration")
	if err := envconfig.Init(cfg); err != nil {
		panic(errors.Wrap(err, "while reading env variables"))
	}

	validationConfig := webhook.ReadValidationConfigOrDie()
	defaultingConfig := webhook.ReadDefaultingConfigOrDie()

	// manager setup
	log.Info("setting up controller-manager")
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		Port:               cfg.WebhookPort,
		MetricsBindAddress: ":9090",
	})
	if err != nil {
		log.Error(err, "failed to setup controller-manager")
		os.Exit(1)
	}

	log.Info("setting up webhook certificates and webhook secret")
	// we need to ensure the certificates and the webhook secret as early as possible
	// because the webhook server needs to read it from disk to start.
	if err := resources.SetupCertificates(
		context.Background(),
		cfg.WebhookSecretName,
		cfg.SystemNamespace,
		cfg.WebhookServiceName); err != nil {
		log.Error(err, "failed to setup certificates and webhook secret")
		os.Exit(1)
	}

	log.Info("setting up webhook server")
	// webhook server setup
	whs := mgr.GetWebhookServer()
	whs.CertName = resources.CertFile
	whs.KeyName = resources.KeyFile

	whs.Register(resources.FunctionConvertWebhookPath,
		webhook.NewConvertingWebHook(
			mgr.GetClient(),
			scheme,
		),
	)

	whs.Register(resources.FunctionDefaultingWebhookPath,
		&ctrlwebhook.Admission{
			Handler: webhook.NewDefaultingWebhook(defaultingConfig, mgr.GetClient()),
		},
	)
	whs.Register(resources.FunctionValidationWebhookPath,
		&ctrlwebhook.Admission{
			Handler: webhook.NewValidatingHook(validationConfig, mgr.GetClient()),
		},
	)

	whs.Register(resources.RegistryConfigDefaultingWebhookPath,
		&ctrlwebhook.Admission{
			Handler: webhook.NewRegistryWatcher(),
		},
	)

	log.Info("setting up webhook resources controller")
	// apply and monitor configuration
	if err := resources.SetupResourcesController(
		context.Background(),
		mgr,
		cfg.WebhookServiceName,
		cfg.SystemNamespace,
		cfg.WebhookSecretName,
	); err != nil {
		log.Error(err, "failed to setup webhook resources controller")
		os.Exit(1)
	}

	log.Info("starting the controller-manager")
	// start the server manager
	err = mgr.Start(ctrl.SetupSignalHandler())
	if err != nil {
		log.Error(err, "failed to start controller-manager")
		os.Exit(1)
	}
}

func setupControllerLogger() {
	atomicLevel := zap.NewAtomicLevelAt(zapcore.DebugLevel)
	zapLogger := ctrlzap.NewRaw(ctrlzap.UseDevMode(true), ctrlzap.Level(&atomicLevel))
	ctrl.SetLogger(zapr.NewLogger(zapLogger))
}
