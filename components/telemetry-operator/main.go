/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"errors"
	"flag"
	"os"
	"strings"
	"time"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	logparsercontroller "github.com/kyma-project/kyma/components/telemetry-operator/controller/logparser"
	logpipelinecontroller "github.com/kyma-project/kyma/components/telemetry-operator/controller/logpipeline"
	tracepipelinereconciler "github.com/kyma-project/kyma/components/telemetry-operator/controller/tracepipeline"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fluentbit/config/builder"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/logger"
	"github.com/kyma-project/kyma/components/telemetry-operator/webhook/dryrun"
	logparserwebhook "github.com/kyma-project/kyma/components/telemetry-operator/webhook/logparser"
	logparservalidation "github.com/kyma-project/kyma/components/telemetry-operator/webhook/logparser/validation"
	logpipelinewebhook "github.com/kyma-project/kyma/components/telemetry-operator/webhook/logpipeline"
	logpipelinevalidation "github.com/kyma-project/kyma/components/telemetry-operator/webhook/logpipeline/validation"

	"github.com/go-logr/zapr"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	k8sWebhook "sigs.k8s.io/controller-runtime/pkg/webhook"
	//+kubebuilder:scaffold:imports
)

var (
	certDir              string
	deniedFilterPlugins  string
	deniedOutputPlugins  string
	enableLeaderElection bool
	enableLogging        bool
	enableTracing        bool
	logFormat            string
	logLevel             string
	metricsAddr          string
	probeAddr            string
	scheme               = runtime.NewScheme()
	setupLog             = ctrl.Log.WithName("setup")
	syncPeriod           time.Duration
	telemetryNamespace   string

	createServiceMonitor         bool
	traceCollectorDeploymentName string
	traceCollectorImage          string

	fluentBitEnvSecret         string
	fluentBitFilesConfigMap    string
	fluentBitPath              string
	fluentBitPluginDirectory   string
	fluentBitInputTag          string
	fluentBitMemoryBufferLimit string
	fluentBitStorageType       string
	fluentBitFsBufferLimit     string
	fluentBitConfigMap         string
	fluentBitSectionsConfigMap string
	fluentBitParsersConfigMap  string
	fluentBitDaemonSet         string
	maxLogPipelines            int
)

//nolint:gochecknoinits
func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(telemetryv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func getEnvOrDefault(envVar string, defaultValue string) string {
	if value, ok := os.LookupEnv(envVar); ok {
		return value
	}
	return defaultValue
}

//+kubebuilder:rbac:groups=telemetry.kyma-project.io,resources=logpipelines,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=telemetry.kyma-project.io,resources=logpipelines/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=telemetry.kyma-project.io,resources=logpipelines/finalizers,verbs=update
//+kubebuilder:rbac:groups=telemetry.kyma-project.io,resources=logparsers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=telemetry.kyma-project.io,resources=logparsers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=telemetry.kyma-project.io,resources=logparsers/finalizers,verbs=update
//+kubebuilder:rbac:groups=telemetry.kyma-project.io,resources=tracepipelines,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=telemetry.kyma-project.io,resources=tracepipelines/status,verbs=get;update;patch

//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete

//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=daemonsets,verbs=get;list;watch;patch

//+kubebuilder:rbac:groups=monitoring.coreos.com,resources=servicemonitors,verbs=get;list;watch;create;update;patch;delete

//+kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=validatingwebhookconfigurations,verbs=get;update;

func main() {
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.DurationVar(&syncPeriod, "sync-period", 1*time.Hour, "minimum frequency at which watched resources are reconciled")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(&enableLogging, "enable-logging", true, "Enable configurable logging.")
	flag.BoolVar(&enableTracing, "enable-tracing", true, "Enable configurable tracing.")
	flag.StringVar(&logFormat, "log-format", getEnvOrDefault("APP_LOG_FORMAT", "text"), "Log format (json or text)")
	flag.StringVar(&logLevel, "log-level", getEnvOrDefault("APP_LOG_LEVEL", "debug"), "Log level (debug, info, warn, error, fatal)")
	flag.StringVar(&certDir, "cert-dir", ".", "Webhook TLS certificate directory")
	flag.StringVar(&telemetryNamespace, "telemetry-namespace", "kyma-system", "Telemetry namespace")

	flag.BoolVar(&createServiceMonitor, "create-service-monitor", true, "Create Prometheus ServiceMonitor for opentelemetry-collector")
	flag.StringVar(&traceCollectorDeploymentName, "trace-collector-deployment-name", "telemetry-trace-collector", "Deployment name for tracing OpenTelemetry Collector")
	flag.StringVar(&traceCollectorImage, "trace-collector-image", "otel/opentelemetry-collector-contrib:0.60.0", "Image for tracing OpenTelemetry Collector")

	flag.StringVar(&fluentBitConfigMap, "fluent-bit-cm-name", "telemetry-fluent-bit", "ConfigMap name of Fluent Bit")
	flag.StringVar(&fluentBitSectionsConfigMap, "fluent-bit-sections-cm-name", "telemetry-fluent-bit-sections", "ConfigMap name of Fluent Bit Sections to be written by Fluent Bit controller")
	flag.StringVar(&fluentBitParsersConfigMap, "fluent-bit-parser-cm-name", "telemetry-fluent-bit-parsers", "ConfigMap name of Fluent Bit Parsers to be written by Fluent Bit controller")
	flag.StringVar(&fluentBitDaemonSet, "fluent-bit-ds-name", "telemetry-fluent-bit", "DaemonSet name to be managed by Fluent Bit controller")
	flag.StringVar(&fluentBitEnvSecret, "fluent-bit-env-secret", "telemetry-fluent-bit-env", "Secret for environment variables")
	flag.StringVar(&fluentBitFilesConfigMap, "fluent-bit-files-cm", "telemetry-fluent-bit-files", "ConfigMap for referenced files")
	flag.StringVar(&fluentBitPath, "fluent-bit-path", "fluent-bit/bin/fluent-bit", "Fluent Bit binary path")
	flag.StringVar(&fluentBitPluginDirectory, "fluent-bit-plugin-directory", "fluent-bit/lib", "Fluent Bit plugin directory")
	flag.StringVar(&fluentBitInputTag, "fluent-bit-input-tag", "tele", "Fluent Bit base tag of the input to use")
	flag.StringVar(&fluentBitMemoryBufferLimit, "fluent-bit-memory-buffer-limit", "10M", "Fluent Bit memory buffer limit per log pipeline")
	flag.StringVar(&fluentBitStorageType, "fluent-bit-storage-type", "filesystem", "Fluent Bit buffering mechanism (filesystem or memory)")
	flag.StringVar(&fluentBitFsBufferLimit, "fluent-bit-filesystem-buffer-limit", "1G", "Fluent Bit filesystem buffer limit per log pipeline")
	flag.StringVar(&deniedFilterPlugins, "fluent-bit-denied-filter-plugins", "", "Comma separated list of denied filter plugins even if allowUnsupportedPlugins is enabled. If empty, all filter plugins are allowed.")
	flag.StringVar(&deniedOutputPlugins, "fluent-bit-denied-output-plugins", "", "Comma separated list of denied output plugins even if allowUnsupportedPlugins is enabled. If empty, all output plugins are allowed.")
	flag.IntVar(&maxLogPipelines, "fluent-bit-max-pipelines", 5, "Maximum number of LogPipelines to be created. If 0, no limit is applied.")

	flag.Parse()
	if err := validateFlags(); err != nil {
		setupLog.Error(err, "Invalid flag provided")
		os.Exit(1)
	}

	ctrLogger, err := logger.New(logFormat, logLevel)
	ctrl.SetLogger(zapr.NewLogger(ctrLogger.WithContext().Desugar()))
	if err != nil {
		setupLog.Error(err, "Failed to initialize logger")
		os.Exit(1)
	}
	defer func() {
		if err := ctrLogger.WithContext().Sync(); err != nil {
			setupLog.Error(err, "Failed to flush logger")
		}
	}()

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		SyncPeriod:             &syncPeriod,
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "cdd7ef0b.kyma-project.io",
		CertDir:                certDir,
	})
	if err != nil {
		setupLog.Error(err, "Failed to start manager")
		os.Exit(1)
	}

	if enableLogging {
		setupLog.Info("Starting with logging controllers")
		mgr.GetWebhookServer().Register("/validate-logpipeline", &k8sWebhook.Admission{Handler: createLogPipelineValidator(mgr.GetClient())})
		mgr.GetWebhookServer().Register("/validate-logparser", &k8sWebhook.Admission{Handler: createLogParserValidator(mgr.GetClient())})

		if err = createLogPipelineReconciler(mgr.GetClient()).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "Failed to create controller", "controller", "LogPipeline")
			os.Exit(1)
		}

		if err = createLogParserReconciler(mgr.GetClient()).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "Failed to create controller", "controller", "LogParser")
			os.Exit(1)
		}
	}

	if enableTracing {
		setupLog.Info("Starting with tracing controller")
		if err = createTracePipelineReconciler(mgr.GetClient()).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "Failed to create controller", "controller", "TracePipeline")
			os.Exit(1)
		}
		if err = monitoringv1.AddToScheme(scheme); err != nil {
			setupLog.Error(err, "Failed to add monitoring scheme", "controller", "TracePipeline")
			os.Exit(1)
		}
	}

	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "Failed to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "Failed to set up ready check")
		os.Exit(1)
	}

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "Failed to run manager")
		os.Exit(1)
	}
}

func validateFlags() error {
	if fluentBitConfigMap == "" {
		return errors.New("--fluent-bit-cm-name flag is required")
	}
	if fluentBitSectionsConfigMap == "" {
		return errors.New("--fluent-bit-sections-cm-name flag is required")
	}
	if fluentBitParsersConfigMap == "" {
		return errors.New("--fluent-bit-parser-cm-name flag is required")
	}
	if fluentBitDaemonSet == "" {
		return errors.New("--fluent-bit-ds-name flag is required")
	}
	if fluentBitEnvSecret == "" {
		return errors.New("--fluent-bit-env-secret flag is required")
	}
	if fluentBitFilesConfigMap == "" {
		return errors.New("--fluent-bit-files-cm flag is required")
	}
	if telemetryNamespace == "" {
		return errors.New("--fluent-bit-ns flag is required")
	}
	if logFormat != "json" && logFormat != "text" {
		return errors.New("--log-format has to be either json or text")
	}
	if logLevel != "debug" && logLevel != "info" && logLevel != "warn" && logLevel != "error" && logLevel != "fatal" {
		return errors.New("--log-level has to be one of debug, info, warn, error, fatal")
	}
	if fluentBitStorageType != "filesystem" && fluentBitStorageType != "memory" {
		return errors.New("--fluent-bit-storage-type has to be either filesystem or memory")
	}
	return nil
}

func createLogPipelineReconciler(client client.Client) *logpipelinecontroller.Reconciler {
	config := logpipelinecontroller.Config{
		SectionsConfigMap: types.NamespacedName{Name: fluentBitSectionsConfigMap, Namespace: telemetryNamespace},
		FilesConfigMap:    types.NamespacedName{Name: fluentBitFilesConfigMap, Namespace: telemetryNamespace},
		EnvSecret:         types.NamespacedName{Name: fluentBitEnvSecret, Namespace: telemetryNamespace},
		DaemonSet:         types.NamespacedName{Namespace: telemetryNamespace, Name: fluentBitDaemonSet},
		PipelineDefaults:  createPipelineDefaults(),
	}
	return logpipelinecontroller.NewReconciler(client, config)
}

func createLogParserReconciler(client client.Client) *logparsercontroller.Reconciler {
	config := logparsercontroller.Config{
		ParsersConfigMap: types.NamespacedName{Name: fluentBitParsersConfigMap, Namespace: telemetryNamespace},
		DaemonSet:        types.NamespacedName{Namespace: telemetryNamespace, Name: fluentBitDaemonSet},
	}
	return logparsercontroller.NewReconciler(client, config)
}

func createLogPipelineValidator(client client.Client) *logpipelinewebhook.ValidatingWebhookHandler {
	return logpipelinewebhook.NewValidatingWebhookHandler(
		client,
		logpipelinevalidation.NewInputValidator(),
		logpipelinevalidation.NewVariablesValidator(client),
		logpipelinevalidation.NewFilterValidator(parsePlugins(deniedFilterPlugins)...),
		logpipelinevalidation.NewMaxPipelinesValidator(maxLogPipelines),
		logpipelinevalidation.NewOutputValidator(parsePlugins(deniedOutputPlugins)...),
		logpipelinevalidation.NewFilesValidator(),
		dryrun.NewDryRunner(client, createDryRunConfig()))
}

func createLogParserValidator(client client.Client) *logparserwebhook.ValidatingWebhookHandler {
	return logparserwebhook.NewValidatingWebhookHandler(
		client,
		logparservalidation.NewParserValidator(),
		dryrun.NewDryRunner(client, createDryRunConfig()))
}

func createTracePipelineReconciler(client client.Client) *tracepipelinereconciler.Reconciler {
	config := tracepipelinereconciler.Config{
		CreateServiceMonitor: createServiceMonitor,
		CollectorNamespace:   telemetryNamespace,
		ResourceName:         traceCollectorDeploymentName,
		CollectorImage:       traceCollectorImage,
	}
	return tracepipelinereconciler.NewReconciler(client, config, scheme)
}

func createDryRunConfig() dryrun.Config {
	return dryrun.Config{
		FluentBitBinPath:       fluentBitPath,
		FluentBitPluginDir:     fluentBitPluginDirectory,
		FluentBitConfigMapName: types.NamespacedName{Name: fluentBitConfigMap, Namespace: telemetryNamespace},
		PipelineDefaults:       createPipelineDefaults(),
	}
}

func createPipelineDefaults() builder.PipelineDefaults {
	return builder.PipelineDefaults{
		InputTag:          fluentBitInputTag,
		MemoryBufferLimit: fluentBitMemoryBufferLimit,
		StorageType:       fluentBitStorageType,
		FsBufferLimit:     fluentBitFsBufferLimit,
	}
}

func parsePlugins(s string) []string {
	return strings.SplitN(strings.ReplaceAll(s, " ", ""), ",", len(s))
}
