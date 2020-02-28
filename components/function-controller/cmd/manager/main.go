package manager

import (
	"flag"
	"os"

	"github.com/kelseyhightower/envconfig"
	"github.com/kyma-project/kyma/components/function-controller/internal/controllers"
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
	Function  controllers.FunctionConfig
	Namespace controllers.NamespaceConfig
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

	cfg, err := loadConfig("APP")
	if err != nil {
		setupLog.Error(err, "unable to load config")
		os.Exit(1)
	}

	restConfig := ctrl.GetConfigOrDie()
	mgr, err := ctrl.NewManager(restConfig, ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: envs.metricsAddr,
		LeaderElection:     envs.enableLeaderElection,
		LeaderElectionID:   envs.leaderElectionID,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	container := &controllers.Container{
		Manager: mgr,
	}

	runControllers(cfg, container, mgr)

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
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

func loadConfig(prefix string) (Config, error) {
	cfg := Config{}
	err := envconfig.Process(prefix, &cfg)
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}

func runControllers(config Config, container *controllers.Container, mgr manager.Manager) {
	controllers := map[string]func(Config, *controllers.Container, manager.Manager) error{
		"Function":  runFunctionController,
		"Namespace": runNamespaceController,
	}

	for name, controller := range controllers {
		err := controller(config, container, mgr)
		if err != nil {
			setupLog.Error(err, "unable to create controller", "controller", name)
			os.Exit(1)
		}
	}
}

func runFunctionController(config Config, container *controllers.Container, mgr manager.Manager) error {
	return controllers.NewFunction(config.Function, ctrl.Log.WithName("controllers").WithName("Function"), container).SetupWithManager(mgr)
}

func runNamespaceController(config Config, container *controllers.Container, mgr manager.Manager) error {
	if !config.Namespace.EnableController {
		return nil
	}

	return controllers.NewNamespace(config.Namespace, ctrl.Log.WithName("controllers").WithName("Function"), container).SetupWithManager(mgr)
}
