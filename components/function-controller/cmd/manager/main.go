package main

import (
	"os"

	"github.com/vrischmann/envconfig"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/kyma-project/kyma/components/function-controller/internal/configwatcher"
	"github.com/kyma-project/kyma/components/function-controller/internal/controllers/knative"
	k8s "github.com/kyma-project/kyma/components/function-controller/internal/controllers/kubernetes"
	"github.com/kyma-project/kyma/components/function-controller/internal/controllers/serverless"
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = servingv1.AddToScheme(scheme)

	_ = serverlessv1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

type config struct {
	MetricsAddress        string `envconfig:"default=:8080"`
	LeaderElectionEnabled bool   `envconfig:"default=false"`
	LeaderElectionID      string `envconfig:"default=serverless-controller-leader-election-helper"`
	ConfigWatcher         configwatcher.Config
	Function              serverless.FunctionConfig
	KService              knative.ServiceConfig
}

func main() {
	ctrl.SetLogger(zap.New())

	config, err := loadConfig("APP")
	if err != nil {
		setupLog.Error(err, "unable to load config")
		os.Exit(1)
	}

	setupLog.Info("Generating Kubernetes client config")
	restConfig := ctrl.GetConfigOrDie()

	setupLog.Info("Initializing controller manager")
	mgr, err := manager.New(restConfig, manager.Options{
		Scheme:             scheme,
		MetricsBindAddress: config.MetricsAddress,
		LeaderElection:     config.LeaderElectionEnabled,
		LeaderElectionID:   config.LeaderElectionID,
	})
	if err != nil {
		setupLog.Error(err, "Unable to initialize controller manager")
		os.Exit(1)
	}

	coreClient, err := corev1.NewForConfig(restConfig)
	if err != nil {
		setupLog.Error(err, "Unable to initialize core client")
		os.Exit(1)
	}

	resourceConfigServices := configwatcher.NewConfigWatcherServices(coreClient, config.ConfigWatcher)

	if err := serverless.NewFunction(mgr.GetClient(), ctrl.Log.WithName("controllers").WithName("function"), config.Function, scheme, mgr.GetEventRecorderFor("function-controller")).
		SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create function controller")
		os.Exit(1)
	}

	if err := knative.NewServiceReconciler(mgr.GetClient(), ctrl.Log.WithName("controllers").WithName("kservice"), config.KService, scheme, mgr.GetEventRecorderFor("kservice-controller")).
		SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create knative service controller")
		os.Exit(1)
	}

	if err := k8s.NewController(mgr.GetClient(), ctrl.Log.WithName("controllers").WithName("namespace"), config.ConfigWatcher, k8s.NamespaceType, resourceConfigServices).
		SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create Namespace controller")
		os.Exit(1)
	}
	if err := k8s.NewController(mgr.GetClient(), ctrl.Log.WithName("controllers").WithName("secret"), config.ConfigWatcher, k8s.SecretType, resourceConfigServices).
		SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create Secret controller")
		os.Exit(1)
	}
	if err := k8s.NewController(mgr.GetClient(), ctrl.Log.WithName("controllers").WithName("configmap"), config.ConfigWatcher, k8s.ConfigMapType, resourceConfigServices).
		SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create ConfigMap controller")
		os.Exit(1)
	}
	if err := k8s.NewController(mgr.GetClient(), ctrl.Log.WithName("controllers").WithName("serviceaccount"), config.ConfigWatcher, k8s.ServiceAccountType, resourceConfigServices).
		SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create ServiceAccount controller")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("Running manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "Unable to run the manager")
		os.Exit(1)
	}
}

func loadConfig(prefix string) (config, error) {
	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, prefix)
	if err != nil {
		return cfg, err
	}

	cfg.Function.Build.LimitsCPUValue = resource.MustParse(cfg.Function.Build.LimitsCPU)
	cfg.Function.Build.LimitsMemoryValue = resource.MustParse(cfg.Function.Build.LimitsMemory)
	cfg.Function.Build.RequestsCPUValue = resource.MustParse(cfg.Function.Build.RequestsCPU)
	cfg.Function.Build.RequestsMemoryValue = resource.MustParse(cfg.Function.Build.RequestsMemory)

	return cfg, nil
}
