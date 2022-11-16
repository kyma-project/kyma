package main

import (
	"context"
	"os"

	"github.com/go-logr/zapr"
	fileconfig "github.com/kyma-project/kyma/components/function-controller/internal/config"
	"github.com/kyma-project/kyma/components/function-controller/internal/logging"
	"github.com/kyma-project/kyma/components/function-controller/internal/webhook"
	"github.com/kyma-project/kyma/components/function-controller/internal/webhook/resources"
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"sigs.k8s.io/controller-runtime/pkg/manager"

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
	WebhookConfigPath  string `envconfig:"default=/appdata/config.yaml"`
}

var (
	scheme = runtime.NewScheme()
)

// nolint
func init() {
	_ = serverlessv1alpha2.AddToScheme(scheme)
	_ = serverlessv1alpha1.AddToScheme(scheme)
	_ = admissionregistrationv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	setupLog := ctrlzap.New().WithName("setup")

	setupLog.Info("reading configuration")
	cfg := &config{}
	if err := envconfig.Init(cfg); err != nil {
		panic(errors.Wrap(err, "while reading env variables"))
	}

	logCfg, err := fileconfig.Load(cfg.WebhookConfigPath)
	if err != nil {
		setupLog.Error(err, "unable to load configuration file")
		os.Exit(1)
	}

	atomic := zap.NewAtomicLevel()
	parsedLevel, err := zapcore.ParseLevel(logCfg.LogLevel)
	if err != nil {
		setupLog.Error(err, "unable to parse logger level")
		os.Exit(1)
	}
	atomic.SetLevel(parsedLevel)

	log, err := logging.ConfigureLogger(logCfg.LogLevel, logCfg.LogFormat, atomic)
	if err != nil {
		setupLog.Error(err, "unable to configure log")
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logWithCtx := log.WithContext()
	go logging.ReconfigureOnConfigChange(ctx, logWithCtx.Named("notifier"), atomic, cfg.WebhookConfigPath)

	logrZap := zapr.NewLogger(logWithCtx.Desugar())
	ctrl.SetLogger(logrZap)

	validationConfigv1alpha1 := webhook.ReadValidationConfigV1Alpha1OrDie()
	validationConfigv1alpha2 := webhook.ReadValidationConfigV1Alpha2OrDie()
	defaultingConfigv1alpha1 := webhook.ReadDefaultingConfigV1Alpha1OrDie()
	defaultingConfigv1alpha2 := webhook.ReadDefaultingConfigV1Alpha2OrDie()

	// manager setup
	logWithCtx.Info("setting up controller-manager")

	mgr, err := manager.New(ctrl.GetConfigOrDie(), manager.Options{
		Scheme:             scheme,
		Port:               cfg.WebhookPort,
		MetricsBindAddress: ":9090",
		Logger:             logrZap,
	})
	if err != nil {
		logWithCtx.Error(err, "failed to setup controller-manager")
		os.Exit(1)
	}

	logWithCtx.Info("setting up webhook certificates and webhook secret")
	// we need to ensure the certificates and the webhook secret as early as possible
	// because the webhook server needs to read it from disk to start.
	if err := resources.SetupCertificates(
		context.Background(),
		cfg.WebhookSecretName,
		cfg.SystemNamespace,
		cfg.WebhookServiceName,
		logWithCtx.Named("setup-certificates")); err != nil {
		logWithCtx.Error(err, "failed to setup certificates and webhook secret")
		os.Exit(1)
	}

	logWithCtx.Info("setting up webhook server")
	// webhook server setup
	whs := mgr.GetWebhookServer()
	whs.CertName = resources.CertFile
	whs.KeyName = resources.KeyFile
	whs.Register(resources.FunctionConvertWebhookPath,
		webhook.NewConvertingWebhook(
			mgr.GetClient(),
			scheme,
			logWithCtx.Named("converting-webhook")),
	)
	whs.Register(resources.FunctionDefaultingWebhookPath, &ctrlwebhook.Admission{
		Handler: webhook.NewDefaultingWebhook(defaultingConfigv1alpha1, defaultingConfigv1alpha2, mgr.GetClient()),
	})

	whs.Register(resources.FunctionValidationWebhookPath, &ctrlwebhook.Admission{
		Handler: webhook.NewValidatingHook(validationConfigv1alpha1, validationConfigv1alpha2, mgr.GetClient()),
	})

	whs.Register(resources.RegistryConfigDefaultingWebhookPath, &ctrlwebhook.Admission{Handler: webhook.NewRegistryWatcher()})

	logWithCtx.Info("setting up webhook resources controller")
	// apply and monitor configuration
	if err := resources.SetupResourcesController(
		context.Background(),
		mgr,
		cfg.WebhookServiceName,
		cfg.SystemNamespace,
		cfg.WebhookSecretName,
		logWithCtx); err != nil {
		logWithCtx.Error(err, "failed to setup webhook resources controller")
		os.Exit(1)
	}

	logWithCtx.Info("starting the controller-manager")
	// start the server manager
	err = mgr.Start(ctrl.SetupSignalHandler())
	if err != nil {
		logWithCtx.Error(err, "failed to start controller-manager")
		os.Exit(1)
	}
}
