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
	"context"
	"errors"
	"flag"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"time"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	logparsercontroller "github.com/kyma-project/kyma/components/telemetry-operator/controller/logparser"
	logpipelinecontroller "github.com/kyma-project/kyma/components/telemetry-operator/controller/logpipeline"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fluentbit/config/builder"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/logger"
	"github.com/kyma-project/kyma/components/telemetry-operator/webhook/dryrun"
	logparserwebhook "github.com/kyma-project/kyma/components/telemetry-operator/webhook/logparser"
	logparservalidation "github.com/kyma-project/kyma/components/telemetry-operator/webhook/logparser/validation"
	logpipelinewebhook "github.com/kyma-project/kyma/components/telemetry-operator/webhook/logpipeline"
	logpipelinevalidation "github.com/kyma-project/kyma/components/telemetry-operator/webhook/logpipeline/validation"

	"github.com/prometheus/client_golang/prometheus"

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
	"sigs.k8s.io/controller-runtime/pkg/metrics"
	k8sWebhook "sigs.k8s.io/controller-runtime/pkg/webhook"
	//+kubebuilder:scaffold:imports
)

var (
	scheme                     = runtime.NewScheme()
	setupLog                   = ctrl.Log.WithName("setup")
	fluentBitConfigMap         string
	fluentBitSectionsConfigMap string
	fluentBitParsersConfigMap  string
	fluentBitDaemonSet         string
	fluentBitNs                string
	fluentBitEnvSecret         string
	fluentBitFilesConfigMap    string
	fluentBitPath              string
	fluentBitPluginDirectory   string
	fluentBitInputTag          string
	fluentBitMemoryBufferLimit string
	fluentBitStorageType       string
	fluentBitFsBufferLimit     string
	logFormat                  string
	logLevel                   string
	certDir                    string
	deniedFilterPlugins        string
	deniedOutputPlugins        string
	maxPipelines               int
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

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var syncPeriod time.Duration
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.DurationVar(&syncPeriod, "sync-period", 1*time.Hour, "minimum frequency at which watched resources are reconciled")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&fluentBitConfigMap, "cm-name", "", "ConfigMap name of Fluent Bit")
	flag.StringVar(&fluentBitSectionsConfigMap, "sections-cm-name", "", "ConfigMap name of Fluent Bit Sections to be written by Fluent Bit controller")
	flag.StringVar(&fluentBitParsersConfigMap, "parser-cm-name", "", "ConfigMap name of Fluent Bit Parsers to be written by Fluent Bit controller")
	flag.StringVar(&fluentBitDaemonSet, "ds-name", "", "DaemonSet name to be managed by Fluent Bit controller")
	flag.StringVar(&fluentBitEnvSecret, "env-secret", "", "Secret for environment variables")
	flag.StringVar(&fluentBitFilesConfigMap, "files-cm", "", "ConfigMap for referenced files")
	flag.StringVar(&fluentBitNs, "fluent-bit-ns", "", "Fluent Bit namespace")
	flag.StringVar(&fluentBitPath, "fluent-bit-path", "fluent-bit/bin/fluent-bit", "Fluent Bit binary path")
	flag.StringVar(&fluentBitPluginDirectory, "fluent-bit-plugin-directory", "fluent-bit/lib", "Fluent Bit plugin directory")
	flag.StringVar(&fluentBitInputTag, "fluent-bit-input-tag", "tele", "Fluent Bit base tag of the input to use")
	flag.StringVar(&fluentBitMemoryBufferLimit, "fluent-bit-memory-buffer-limit", "10M", "Fluent Bit memory buffer limit per log pipeline")
	flag.StringVar(&fluentBitStorageType, "fluent-bit-storage-type", "filesystem", "Fluent Bit buffering mechanism (filesystem or memory)")
	flag.StringVar(&fluentBitFsBufferLimit, "fluent-bit-filesystem-buffer-limit", "1G", "Fluent Bit filesystem buffer limit per log pipeline")
	flag.StringVar(&logFormat, "log-format", getEnvOrDefault("APP_LOG_FORMAT", "text"), "Log format (json or text)")
	flag.StringVar(&logLevel, "log-level", getEnvOrDefault("APP_LOG_LEVEL", "debug"), "Log level (debug, info, warn, error, fatal)")
	flag.StringVar(&certDir, "cert-dir", "/var/run/telemetry-webhook", "Webhook TLS certificate directory")
	flag.StringVar(&deniedFilterPlugins, "denied-filter-plugins", "", "Comma separated list of denied filter plugins even if allowUnsupportedPlugins is enabled. If empty, all filter plugins are allowed.")
	flag.StringVar(&deniedOutputPlugins, "denied-output-plugins", "", "Comma separated list of denied output plugins even if allowUnsupportedPlugins is enabled. If empty, all output plugins are allowed.")
	flag.IntVar(&maxPipelines, "max-pipelines", 5, "Maximum number of LogPipelines to be created. If 0, no limit is applied.")

	flag.Parse()

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

	if err := validateFlags(); err != nil {
		setupLog.Error(err, "Invalid flag provided")
		os.Exit(1)
	}

	restartsTotal := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "telemetry_fluentbit_triggered_restarts_total",
		Help: "Number of triggered Fluent Bit restarts",
	})
	metrics.Registry.MustRegister(restartsTotal)

	if err != nil {
		setupLog.Error(err, "Failed to set watch")
		os.Exit(1)
	}

	go watcher()

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		SyncPeriod:             &syncPeriod,
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "cdd7ef0b.kyma-project.io",
		CertDir:                certDir,
		//NewCache: cache.BuilderWithOptions(cache.Options{
		//	Scheme: scheme,
		//	SelectorsByObject: cache.SelectorsByObject{
		//		&corev1.Secret{}: {
		//			Field: fields.SelectorFromSet(fields.Set{}),
		//		},
		//	},
		//}),
	})
	if err != nil {
		setupLog.Error(err, "Failed to start manager")
		os.Exit(1)
	}

	pipelineDefaults := builder.PipelineDefaults{
		InputTag:          fluentBitInputTag,
		MemoryBufferLimit: fluentBitMemoryBufferLimit,
		StorageType:       fluentBitStorageType,
		FsBufferLimit:     fluentBitFsBufferLimit,
	}

	daemonSet := types.NamespacedName{
		Namespace: fluentBitNs,
		Name:      fluentBitDaemonSet,
	}
	logpipelineConfig := logpipelinecontroller.Config{
		SectionsConfigMap: types.NamespacedName{
			Name:      fluentBitSectionsConfigMap,
			Namespace: fluentBitNs,
		},

		FilesConfigMap: types.NamespacedName{
			Name:      fluentBitFilesConfigMap,
			Namespace: fluentBitNs,
		},
		EnvSecret: types.NamespacedName{
			Name:      fluentBitEnvSecret,
			Namespace: fluentBitNs,
		},
		DaemonSet:        daemonSet,
		PipelineDefaults: pipelineDefaults,
	}

	logparserConfig := logparsercontroller.Config{
		ParsersConfigMap: types.NamespacedName{
			Name:      fluentBitParsersConfigMap,
			Namespace: fluentBitNs,
		},
		DaemonSet: daemonSet,
	}

	dryRunConfig := dryrun.Config{
		FluentBitBinPath:       fluentBitPath,
		FluentBitPluginDir:     fluentBitPluginDirectory,
		FluentBitConfigMapName: types.NamespacedName{Name: fluentBitConfigMap, Namespace: fluentBitNs},
		PipelineDefaults:       pipelineDefaults,
	}

	logPipelineValidationHandler := logpipelinewebhook.NewValidatingWebhookHandler(
		mgr.GetClient(),
		logpipelinevalidation.NewInputValidator(),
		logpipelinevalidation.NewVariablesValidator(mgr.GetClient()),
		logpipelinevalidation.NewFilterValidator(parsePlugins(deniedFilterPlugins)...),
		logpipelinevalidation.NewMaxPipelinesValidator(maxPipelines),
		logpipelinevalidation.NewOutputValidator(parsePlugins(deniedOutputPlugins)...),
		logpipelinevalidation.NewFilesValidator(),
		dryrun.NewDryRunner(mgr.GetClient(), dryRunConfig))

	logParserValidationHandler := logparserwebhook.NewValidatingWebhookHandler(
		mgr.GetClient(),
		logparservalidation.NewParserValidator(),
		dryrun.NewDryRunner(mgr.GetClient(), dryRunConfig))

	mgr.GetWebhookServer().Register(
		"/validate-logpipeline",
		&k8sWebhook.Admission{Handler: logPipelineValidationHandler})
	mgr.GetWebhookServer().Register(
		"/validate-logparser",
		&k8sWebhook.Admission{Handler: logParserValidationHandler})

	logPipelineReconciler := logpipelinecontroller.NewReconciler(
		mgr.GetClient(),
		logpipelineConfig,
		restartsTotal)
	if err = logPipelineReconciler.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "Failed to create controller", "controller", "LogPipeline")
		os.Exit(1)
	}

	logParserReconciler := logparsercontroller.NewReconciler(
		mgr.GetClient(),
		logparserConfig,
		restartsTotal)
	if err = logParserReconciler.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "Failed to create controller", "controller", "LogParser")
		os.Exit(1)
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
		return errors.New("--cm-name flag is required")
	}
	if fluentBitSectionsConfigMap == "" {
		return errors.New("--sections-cm-name flag is required")
	}
	if fluentBitParsersConfigMap == "" {
		return errors.New("--parser-cm-name flag is required")
	}
	if fluentBitDaemonSet == "" {
		return errors.New("--ds-name flag is required")
	}
	if fluentBitEnvSecret == "" {
		return errors.New("--env-secret flag is required")
	}
	if fluentBitFilesConfigMap == "" {
		return errors.New("--files-cm flag is required")
	}
	if fluentBitNs == "" {
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

func parsePlugins(s string) []string {
	return strings.SplitN(strings.ReplaceAll(s, " ", ""), ",", len(s))
}

func watcher() {
	cl, err := client.NewWithWatch(ctrl.GetConfigOrDie(), client.Options{Scheme: scheme})
	if err != nil {
		setupLog.Error(err, "Failed to set up health check")
		os.Exit(1)
	}
	sec1 := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mysecret",
			Namespace: "cls",
		},
	}
	secretList := &corev1.SecretList{}
	secretList.Items = append(secretList.Items, sec1)
	fmt.Printf("secretlist: %+v\n", secretList)
	w, err := cl.Watch(context.TODO(), secretList)
	for {
		event, ok := <-w.ResultChan()
		if ok {
			metaObject, ok := event.Object.(metav1.Object)
			if ok {
				fmt.Printf("watching secret name:%s\n", metaObject.GetName())
			}
		}
	}

}
