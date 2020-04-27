package main

import (
	"os"

	"github.com/vrischmann/envconfig"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/kyma-project/kyma/components/function-controller/internal/controllers/knative"
	k8s "github.com/kyma-project/kyma/components/function-controller/internal/controllers/kubernetes"
	"github.com/kyma-project/kyma/components/function-controller/internal/controllers/serverless"
	internalresource "github.com/kyma-project/kyma/components/function-controller/internal/resource"
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
	Kubernetes            k8s.Config
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

	resourceClient := internalresource.New(mgr.GetClient(), scheme)
	configMapSvc := k8s.NewConfigMapService(resourceClient, config.Kubernetes)
	secretSvc := k8s.NewSecretService(resourceClient, config.Kubernetes)
	serviceAccountSvc := k8s.NewServiceAccountService(resourceClient, config.Kubernetes)

	if err := serverless.NewFunction(resourceClient, ctrl.Log, config.Function, mgr.GetEventRecorderFor("function-controller")).
		SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create Function controller")
		os.Exit(1)
	}

	if err := knative.NewServiceReconciler(resourceClient, ctrl.Log, config.KService).
		SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create Knative Service controller")
		os.Exit(1)
	}

	if err := k8s.NewConfigMap(mgr.GetClient(), ctrl.Log, config.Kubernetes, configMapSvc).
		SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create ConfigMap controller")
		os.Exit(1)
	}

	if err := k8s.NewNamespace(mgr.GetClient(), ctrl.Log, config.Kubernetes, configMapSvc, secretSvc, serviceAccountSvc).
		SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create Namespace controller")
		os.Exit(1)
	}

	if err := k8s.NewSecret(mgr.GetClient(), ctrl.Log, config.Kubernetes, secretSvc).
		SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create Secret controller")
		os.Exit(1)
	}

	if err := k8s.NewServiceAccount(mgr.GetClient(), ctrl.Log, config.Kubernetes, serviceAccountSvc).
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
