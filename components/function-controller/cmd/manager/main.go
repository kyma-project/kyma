package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/go-logr/zapr"
	k8s "github.com/kyma-project/kyma/components/function-controller/internal/controllers/kubernetes"
	"github.com/kyma-project/kyma/components/function-controller/internal/controllers/serverless"
	"github.com/kyma-project/kyma/components/function-controller/internal/git"
	internalresource "github.com/kyma-project/kyma/components/function-controller/internal/resource"
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/vrischmann/envconfig"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	ctrlzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = serverlessv1alpha1.AddToScheme(scheme)
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
}

type healthzConfig struct {
	Address         string        `envconfig:"default=:8090"`
	LivenessTimeout time.Duration `envconfig:"default=1s"`
}

func main() {
	config, err := loadConfig("APP")
	if err != nil {
		ctrl.SetLogger(ctrlzap.New())
		setupLog.Error(err, "unable to load config")
		os.Exit(1)
	}

	logLevel, err := toZapLogLevel(config.LogLevel)
	if err != nil {
		ctrl.SetLogger(ctrlzap.New())
		setupLog.Error(err, "unable to set logging level")
		os.Exit(2)
	}

	atomicLevel := zap.NewAtomicLevelAt(logLevel)
	zapLogger := ctrlzap.NewRaw(ctrlzap.UseDevMode(true), ctrlzap.Level(&atomicLevel))
	ctrl.SetLogger(zapr.NewLogger(zapLogger))

	setupLog.Info("Generating Kubernetes client config")
	restConfig := ctrl.GetConfigOrDie()

	setupLog.Info("Initializing controller manager")
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

	mgr.GetWebhookServer().Register(
		"/mutate-v1-secret",
		&webhook.Admission{
			Handler: k8s.NewRegistryWatcher(mgr.GetClient()),
		},
	)

	events := make(chan event.GenericEvent)
	healthCh := make(chan bool)
	healthHandler := serverless.NewHealthChecker(events, healthCh, config.Healthz.LivenessTimeout, zapLogger.Named("healthz"))
	if err := mgr.AddHealthzCheck("health check", healthHandler.Checker); err != nil {
		setupLog.Error(err, "unable to register healthz")
		os.Exit(1)
	}

	fnRecon := serverless.NewFunction(resourceClient, ctrl.Log, config.Function, git.NewGit2Go(), mgr.GetEventRecorderFor(serverlessv1alpha1.FunctionControllerValue), healthCh)
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

	if err := k8s.NewConfigMap(mgr.GetClient(), ctrl.Log, config.Kubernetes, configMapSvc).
		SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create ConfigMap controller")
		os.Exit(1)
	}

	if err := k8s.NewNamespace(mgr.GetClient(), ctrl.Log, config.Kubernetes, configMapSvc, secretSvc, serviceAccountSvc, roleSvc, roleBindingSvc).
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

	if err := k8s.NewRole(mgr.GetClient(), ctrl.Log, config.Kubernetes, roleSvc).
		SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create Role controller")
		os.Exit(1)
	}

	if err := k8s.NewRoleBinding(mgr.GetClient(), ctrl.Log, config.Kubernetes, roleBindingSvc).
		SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create RoleBinding controller")
		os.Exit(1)
	}

	// +kubebuilder:scaffold:builder

	setupLog.Info("Running manager")

	metrics.Registry.MustRegister(serverless.ReconcileCounter, serverless.FunctionConfiguredStatusGaugeVec, serverless.FunctionBuiltStatusGaugeVec, serverless.FunctionRunningStatusGaugeVec)

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
		return 0, errors.New(fmt.Sprintf("Desired log level: %s not exist", level))
	}
}
