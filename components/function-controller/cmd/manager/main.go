package main

import (
	"fmt"
	"os"
	"time"

	"github.com/go-logr/zapr"
	k8s "github.com/kyma-project/kyma/components/function-controller/internal/controllers/kubernetes"
	"github.com/kyma-project/kyma/components/function-controller/internal/controllers/serverless"
	"github.com/kyma-project/kyma/components/function-controller/internal/controllers/serverless/metrics"
	"github.com/kyma-project/kyma/components/function-controller/internal/git"
	serverlessLogging "github.com/kyma-project/kyma/components/function-controller/internal/logging"
	internalresource "github.com/kyma-project/kyma/components/function-controller/internal/resource"
	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	"github.com/vrischmann/envconfig"
	"go.uber.org/zap/zapcore"
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

//nolint
func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = serverlessv1alpha2.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

type config struct {
	MetricsAddress            string `envconfig:"default=:8080"`
	Healthz                   healthzConfig
	LeaderElectionEnabled     bool   `envconfig:"default=false"`
	LeaderElectionID          string `envconfig:"default=serverless-controller-leader-election-helper"`
	SecretMutatingWebhookPort int    `envconfig:"default=8443"`
	LogLevel                  string `envconfig:"default=info"`
	Kubernetes                k8s.Config
	Function                  serverless.FunctionConfig
	LogFormat                 string `envconfig:"default=text"`
}

type healthzConfig struct {
	Address         string        `envconfig:"default=:8090"`
	LivenessTimeout time.Duration `envconfig:"default=10s"`
}

func main() {
	config, err := loadConfig("APP")
	if err != nil {
		ctrl.SetLogger(ctrlzap.New())
		setupLog.Error(err, "unable to load config")
		os.Exit(1)
	}
	l, err := serverlessLogging.ConfigureLogger(config.LogLevel, config.LogFormat)
	if err != nil {
		setupLog.Error(err, "unable to configure logger")
		os.Exit(1)
	}

	ctrl.SetLogger(zapr.NewLogger(l.WithContext().Desugar()))

	zapLogger := l.WithContext()
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
	healthHandler := serverless.NewHealthChecker(events, healthCh, config.Healthz.LivenessTimeout, zapLogger.Named("healthz"))
	if err := mgr.AddHealthzCheck("health check", healthHandler.Checker); err != nil {
		setupLog.Error(err, "unable to register healthz")
		os.Exit(1)
	}

	if err := mgr.AddReadyzCheck("readiness check", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to register readyz")
		os.Exit(1)
	}

	fnRecon := serverless.NewFunction(resourceClient, zapLogger, config.Function, git.NewGit2Go(zapLogger), mgr.GetEventRecorderFor(serverlessv1alpha2.FunctionControllerValue), prometheusCollector, healthCh)
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

	if err := k8s.NewConfigMap(mgr.GetClient(), zapLogger, config.Kubernetes, configMapSvc).
		SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create ConfigMap controller")
		os.Exit(1)
	}

	if err := k8s.NewNamespace(mgr.GetClient(), zapLogger, config.Kubernetes, configMapSvc, secretSvc, serviceAccountSvc, roleSvc, roleBindingSvc).
		SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create Namespace controller")
		os.Exit(1)
	}

	if err := k8s.NewSecret(mgr.GetClient(), zapLogger, config.Kubernetes, secretSvc).
		SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create Secret controller")
		os.Exit(1)
	}

	if err := k8s.NewServiceAccount(mgr.GetClient(), zapLogger, config.Kubernetes, serviceAccountSvc).
		SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create ServiceAccount controller")
		os.Exit(1)
	}

	if err := k8s.NewRole(mgr.GetClient(), zapLogger, config.Kubernetes, roleSvc).
		SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create Role controller")
		os.Exit(1)
	}

	if err := k8s.NewRoleBinding(mgr.GetClient(), zapLogger, config.Kubernetes, roleBindingSvc).
		SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create RoleBinding controller")
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

	return cfg, nil
}

func toZapLogLevel(level string) (zapcore.Level, error) {
	switch level {
	case "debug":
		return zapcore.DebugLevel, nil
	case "info":
		return zapcore.InfoLevel, nil
	case "warn":
		return zapcore.WarnLevel, nil
	case "error":
		return zapcore.ErrorLevel, nil
	default:
		return 0, fmt.Errorf("Desired log level: %s not exist", level)
	}
}
