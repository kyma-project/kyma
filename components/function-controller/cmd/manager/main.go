package main

import (
	"context"
	"os"
	"time"

	"github.com/go-logr/zapr"
	fileconfig "github.com/kyma-project/kyma/components/function-controller/internal/config"
	k8s "github.com/kyma-project/kyma/components/function-controller/internal/controllers/kubernetes"
	"github.com/kyma-project/kyma/components/function-controller/internal/controllers/serverless"
	"github.com/kyma-project/kyma/components/function-controller/internal/controllers/serverless/gitrepo"
	"github.com/kyma-project/kyma/components/function-controller/internal/controllers/serverless/metrics"
	"github.com/kyma-project/kyma/components/function-controller/internal/git"
	"github.com/kyma-project/kyma/components/function-controller/internal/logging"
	internalresource "github.com/kyma-project/kyma/components/function-controller/internal/resource"
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	"github.com/vrischmann/envconfig"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	ctrlzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/source"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrlzap.New().WithName("setup")
)

// nolint
func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = serverlessv1alpha2.AddToScheme(scheme)
	_ = serverlessv1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

type config struct {
	MetricsAddress            string `envconfig:"default=:8080"`
	Healthz                   healthzConfig
	LeaderElectionEnabled     bool   `envconfig:"default=false"`
	LeaderElectionID          string `envconfig:"default=serverless-controller-leader-election-helper"`
	SecretMutatingWebhookPort int    `envconfig:"default=8443"`
	Kubernetes                k8s.Config
	Function                  serverless.FunctionConfig
	FileConfigPath            string `envconfig:"default=/appdata/config.yaml"`
}

type healthzConfig struct {
	Address         string        `envconfig:"default=:8090"`
	LivenessTimeout time.Duration `envconfig:"default=10s"`
}

func main() {
	config, err := loadConfig("APP")
	if err != nil {
		setupLog.Error(err, "unable to load config")
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	l, err := configureLogger(ctx, config.FileConfigPath)
	if err != nil {
		setupLog.Error(err, "unable to configure logger")
		os.Exit(1)
	}

	ctrl.SetLogger(zapr.NewLogger(l.CreateDesugared()))

	zapLogger := l.CreateUnregistered()
	zapLogger.Info("Generating Kubernetes client config")
	restConfig := ctrl.GetConfigOrDie()

	zapLogger.Info("Registering Prometheus Stats Collector")
	prometheusCollector := metrics.NewPrometheusStatsCollector()
	prometheusCollector.Register()

	zapLogger.Info("Initializing controller manager")
	mgr, err := manager.New(restConfig, manager.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     config.MetricsAddress,
		LeaderElection:         config.LeaderElectionEnabled,
		LeaderElectionID:       config.LeaderElectionID,
		Port:                   config.SecretMutatingWebhookPort,
		HealthProbeBindAddress: config.Healthz.Address,
	})
	if err != nil {
		setupLog.Error(err, "Unable to initialize controller manager")
		os.Exit(1)
	}

	resourceClient := internalresource.New(mgr.GetClient(), scheme)
	configMapSvc := k8s.NewConfigMapService(resourceClient, config.Kubernetes)
	secretSvc := k8s.NewSecretService(resourceClient, config.Kubernetes)
	serviceAccountSvc := k8s.NewServiceAccountService(resourceClient, config.Kubernetes)
	roleSvc := k8s.NewRoleService(resourceClient, config.Kubernetes)
	roleBindingSvc := k8s.NewRoleBindingService(resourceClient, config.Kubernetes)

	events := make(chan event.GenericEvent)
	healthCh := make(chan bool)
	healthHandler := serverless.NewHealthChecker(events, healthCh, config.Healthz.LivenessTimeout, l.CreateNamed("healthz"))
	if err := mgr.AddHealthzCheck("health check", healthHandler.Checker); err != nil {
		setupLog.Error(err, "unable to register healthz")
		os.Exit(1)
	}

	if err := mgr.AddReadyzCheck("readiness check", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to register readyz")
		os.Exit(1)
	}

	err = gitrepo.NewGitRepoUpdateController(mgr.GetClient(), l.CreateNamed("controllers.gitrepo")).SetupWithManager(mgr)
	if err != nil {
		setupLog.Error(err, "unable to create gitRepo controller")
		os.Exit(1)
	}

	fnRecon := serverless.NewFunctionReconciler(resourceClient, l.CreateNamed("controllers.function"), config.Function, &git.GitClientFactory{}, mgr.GetEventRecorderFor(serverlessv1alpha2.FunctionControllerValue), prometheusCollector, healthCh)
	fnCtrl, err := fnRecon.SetupWithManager(mgr)
	if err != nil {
		setupLog.Error(err, "unable to create Function controller")
		os.Exit(1)
	}

	err = fnCtrl.Watch(&source.Channel{Source: events}, &handler.EnqueueRequestForObject{})
	if err != nil {
		setupLog.Error(err, "unable to watch something")
		os.Exit(1)
	}

	if err := k8s.NewConfigMap(mgr.GetClient(), l.CreateNamed("controllers.configmap"), config.Kubernetes, configMapSvc).
		SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create ConfigMap controller")
		os.Exit(1)
	}

	if err := k8s.NewNamespace(mgr.GetClient(), l.CreateNamed("controllers.namespace"), config.Kubernetes, configMapSvc, secretSvc, serviceAccountSvc, roleSvc, roleBindingSvc).
		SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create Namespace controller")
		os.Exit(1)
	}

	if err := k8s.NewSecret(mgr.GetClient(), l.CreateNamed("controllers.secret"), config.Kubernetes, secretSvc).
		SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create Secret controller")
		os.Exit(1)
	}

	if err := k8s.NewServiceAccount(mgr.GetClient(), l.CreateNamed("controllers.serviceaccount", "kind", "ServiceAccount"), config.Kubernetes, serviceAccountSvc).
		SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create ServiceAccount controller")
		os.Exit(1)
	}

	if err := k8s.NewRole(mgr.GetClient(), l.CreateNamed("controllers.role"), config.Kubernetes, roleSvc).
		SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create Role controller")
		os.Exit(1)
	}

	if err := k8s.NewRoleBinding(mgr.GetClient(), l.CreateNamed("controllers.rolebinding"), config.Kubernetes, roleBindingSvc).
		SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create RoleBinding controller")
		os.Exit(1)
	}

	// +kubebuilder:scaffold:builder

	zapLogger.Info("Running manager")

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "Unable to run the manager")
		os.Exit(1)
	}
}

func configureLogger(ctx context.Context, cfgPath string) (*logging.Registry, error) {
	cfg, err := fileconfig.Load(cfgPath)
	if err != nil {
		return nil, err
	}

	log, err := logging.ConfigureRegisteredLogger(cfg.LogLevel, cfg.LogFormat)

	notifyLog := log.CreateNamed("notifier")
	go fileconfig.RunOnConfigChange(ctx, notifyLog, cfgPath, func(cfg fileconfig.Config) {
		log, err = log.Reconfigure(cfg.LogLevel, cfg.LogFormat)
		if err != nil {
			notifyLog.Error(err)
		}
	})

	return log, err
}

func loadConfig(prefix string) (config, error) {
	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, prefix)
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}
