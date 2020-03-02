package manager

import (
	"flag"
	"os"

	"github.com/kyma-project/kyma/components/function-controller/internal/container"
	"github.com/kyma-project/kyma/components/function-controller/internal/controllers/configmap"
	"github.com/kyma-project/kyma/components/function-controller/internal/controllers/secret"
	resource_watcher "github.com/kyma-project/kyma/components/function-controller/internal/resource-watcher"
	"k8s.io/client-go/dynamic"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/kyma-project/kyma/components/function-controller/internal/controllers/function"
	"github.com/kyma-project/kyma/components/function-controller/internal/controllers/namespace"

	"github.com/kelseyhightower/envconfig"
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = serverlessv1alpha1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

type Config struct {
	Function              function.FunctionConfig
	ResourceWatcherConfig resource_watcher.ResourceWatcherConfig
}

type Envs struct {
	metricsAddr          string
	enableLeaderElection bool
	leaderElectionID     string
	devLog               bool
}

func main() {
	envs := loadEnvs()

	ctrl.SetLogger(zap.New(func(o *zap.Options) {
		o.Development = envs.devLog
	}))

	cfg, err := loadConfig()
	failOnError(err, "unable to load config")

	restConfig := ctrl.GetConfigOrDie()
	mgr, err := ctrl.NewManager(restConfig, ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: envs.metricsAddr,
		LeaderElection:     envs.enableLeaderElection,
		LeaderElectionID:   envs.leaderElectionID,
	})
	failOnError(err, "unable to start manager")

	coreClient, err := v1.NewForConfig(restConfig)
	failOnError(err, "unable to initialize dynamic client")

	dynamicClient, err := dynamic.NewForConfig(restConfig)
	failOnError(err, "unable to initialize dynamic client")

	resourceWatcherServices := resource_watcher.NewResourceWatcherServices(coreClient, cfg.ResourceWatcherConfig)

	container := &container.Container{
		Manager:                 mgr,
		CoreClient:              coreClient,
		DynamicClient:           &dynamicClient,
		ResourceWatcherServices: resourceWatcherServices,
	}

	runControllers(cfg, container, mgr)

	setupLog.Info("starting manager")
	err = mgr.Start(ctrl.SetupSignalHandler())
	failOnError(err, "problem with running manager")
}

func loadEnvs() *Envs {
	var metricsAddr string
	var enableLeaderElection bool
	var leaderElectionID string
	var devLog bool

	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&leaderElectionID, "leader-election-id", "function-controller-leader-election-helper",
		"ID for leader election for controller manager.")
	flag.BoolVar(&devLog, "devlog", false, "Enable logger's development mode")
	flag.Parse()

	return &Envs{
		metricsAddr,
		enableLeaderElection,
		leaderElectionID,
		devLog,
	}
}

func loadConfig() (Config, error) {
	cfg := Config{}
	err := envconfig.Process("APP", &cfg)
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}

func runControllers(config Config, di *container.Container, mgr manager.Manager) {
	controllers := map[string]func(Config, *container.Container, manager.Manager, string) error{
		"Function": runFunctionController,

		// Controllers for resource watcher
		"Namespace": runNamespaceController,
		"Secret":    runSecretController,
		"ConfigMap": runConfigMapController,
	}

	for name, controller := range controllers {
		err := controller(config, di, mgr, name)
		failOnError(err, "unable to create controller", "controller", name)
	}
}

func runFunctionController(config Config, container *container.Container, mgr manager.Manager, name string) error {
	return function.NewController(config.Function, ctrl.Log.WithName("controllers").WithName(name), container).SetupWithManager(mgr)
}

func runNamespaceController(config Config, container *container.Container, mgr manager.Manager, name string) error {
	if !config.ResourceWatcherConfig.EnableControllers {
		return nil
	}

	return namespace.NewController(config.ResourceWatcherConfig, ctrl.Log.WithName("controllers").WithName(name), container).SetupWithManager(mgr)
}

func runSecretController(config Config, container *container.Container, mgr manager.Manager, name string) error {
	if !config.ResourceWatcherConfig.EnableControllers {
		return nil
	}

	return secret.NewController(config.ResourceWatcherConfig, ctrl.Log.WithName("controllers").WithName(name), container).SetupWithManager(mgr)
}

func runConfigMapController(config Config, container *container.Container, mgr manager.Manager, name string) error {
	if !config.ResourceWatcherConfig.EnableControllers {
		return nil
	}

	return configmap.NewController(config.ResourceWatcherConfig, ctrl.Log.WithName("controllers").WithName(name), container).SetupWithManager(mgr)
}

func failOnError(err error, msg string, args ...string) {
	if err != nil {
		setupLog.Error(err, msg, args)
		os.Exit(1)
	}
}
