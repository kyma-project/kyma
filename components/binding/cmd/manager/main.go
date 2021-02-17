package main

import (
	"github.com/kyma-project/kyma/components/binding/internal/controller"
	"github.com/kyma-project/kyma/components/binding/internal/storage"
	"github.com/kyma-project/kyma/components/binding/internal/target"
	"github.com/kyma-project/kyma/components/binding/internal/webhook"
	bindingMutate "github.com/kyma-project/kyma/components/binding/internal/webhook/binding/mutate"
	bindingValidate "github.com/kyma-project/kyma/components/binding/internal/webhook/binding/validate"
	podMutate "github.com/kyma-project/kyma/components/binding/internal/webhook/pod/mutate"
	targetKindValidate "github.com/kyma-project/kyma/components/binding/internal/webhook/targetkind/validate"
	bindingsv1alpha1 "github.com/kyma-project/kyma/components/binding/pkg/apis/v1alpha1"
	"k8s.io/client-go/dynamic"

	log "github.com/sirupsen/logrus"
	"github.com/vrischmann/envconfig"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	k8sWebhook "sigs.k8s.io/controller-runtime/pkg/webhook"
)

type Config struct {
	DebugMode bool `envconfig:"default=false"`

	Port           int    `envconfig:"default=8443"`
	MetricsAddress string `envconfig:"default=:8080"`
}

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = bindingsv1alpha1.AddToScheme(scheme)
}

func main() {
	var cfg Config
	err := envconfig.InitWithPrefix(&cfg, "APP")
	fatalOnError(err, "while initialization config")

	logger := log.New()
	logger.Info("Start Binding component")
	if !cfg.DebugMode {
		logger.SetLevel(log.ErrorLevel)
		logger.SetLevel(log.FatalLevel)
		logger.SetLevel(log.WarnLevel)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: cfg.MetricsAddress,
		Port:               cfg.Port,
		CertDir:            "/var/run/webhook",
	})
	fatalOnError(err, "while creating new manager")

	targetKindStorage := storage.NewKindStorage()
	dc, err := dynamic.NewForConfig(ctrl.GetConfigOrDie())
	fatalOnError(err, "while creating dynamic client")

	webhookClient := webhook.NewClient(mgr.GetClient())
	mgr.GetWebhookServer().Register(
		"/pod-mutating",
		&k8sWebhook.Admission{Handler: podMutate.NewMutationHandler(webhookClient, log.WithField("webhook", "pod-mutating"))})
	mgr.GetWebhookServer().Register(
		"/binding-mutating",
		&k8sWebhook.Admission{Handler: bindingMutate.NewMutationHandler(targetKindStorage, webhookClient, dc, log.WithField("webhook", "binding-mutating"))})
	mgr.GetWebhookServer().Register(
		"/binding-validating",
		&k8sWebhook.Admission{Handler: bindingValidate.NewValidationHandler(log.WithField("webhook", "binding-validating"))})
	mgr.GetWebhookServer().Register(
		"/targetkind-validating",
		&k8sWebhook.Admission{Handler: targetKindValidate.NewValidationHandler(log.WithField("webhook", "targetkind-validating"))})

	targetKindManager := target.NewHandler(dc, targetKindStorage)

	targetKindReconciler := controller.SetupTargetKindReconciler(mgr.GetClient(), dc, logger, targetKindStorage, mgr.GetScheme())
	fatalOnError(targetKindReconciler.SetupWithManager(mgr), "while creating TargetKindReconciler")

	//TODO: wait for all TargetKind to be synced and registered

	bindingReconciler := controller.SetupBindingReconciler(mgr.GetClient(), targetKindManager, logger, mgr.GetScheme())
	fatalOnError(bindingReconciler.SetupWithManager(mgr), "while creating BindingReconciler")

	fatalOnError(mgr.Start(ctrl.SetupSignalHandler()), "unable to run the manager")
}

func fatalOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err.Error())
	}
}
