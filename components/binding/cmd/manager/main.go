package main

import (
	"github.com/kyma-project/kyma/components/binding/internal/controller"
	"github.com/kyma-project/kyma/components/binding/internal/webhook"
	bindingsv1alpha1 "github.com/kyma-project/kyma/components/binding/pkg/apis/v1alpha1"

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
		CertDir:            "/var/run/binding-pod-webhook",
	})
	fatalOnError(err, "while creating new manager")

	mgr.GetWebhookServer().Register(
		"/binding-mutating-pod",
		&k8sWebhook.Admission{Handler: webhook.NewPodHandler(mgr.GetClient(), log.WithField("webhook", "Pod"))})

	bindingReconciler := controller.SetupBindingReconciler(mgr.GetClient(), logger, mgr.GetScheme())
	fatalOnError(bindingReconciler.SetupWithManager(mgr), "while creating BindingReconciler")

	targetKindReconciler := controller.SetupTargetKindReconciler(mgr.GetClient(), logger, mgr.GetScheme())
	fatalOnError(targetKindReconciler.SetupWithManager(mgr), "while creating TargetKindReconciler")

	fatalOnError(mgr.Start(ctrl.SetupSignalHandler()), "unable to run the manager")
}

func fatalOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err.Error())
	}
}
