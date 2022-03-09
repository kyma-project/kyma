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

	"github.com/go-logr/zapr"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/api/v1alpha1"
	"github.com/kyma-project/kyma/components/telemetry-operator/controllers"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/logger"
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
	logFormat                  string
	logLevel                   string
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
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&fluentBitConfigMap, "cm-name", "", "ConfigMap name of Fluent Bit")
	flag.StringVar(&fluentBitSectionsConfigMap, "sections-cm-name", "", "ConfigMap name of Fluent Bit sections to be written by Fluent Bit controller")
	flag.StringVar(&fluentBitParsersConfigMap, "parser-cm-name", "", "ConfigMap name of Fluent Bit Parsers to be written by Fluent Bit controller")
	flag.StringVar(&fluentBitDaemonSet, "ds-name", "", "DaemonSet name to be managed by Fluent Bit controller")
	flag.StringVar(&fluentBitEnvSecret, "env-secret", "", "Secret for environment variables")
	flag.StringVar(&fluentBitFilesConfigMap, "files-cm", "", "ConfigMap for referenced files")
	flag.StringVar(&fluentBitNs, "fluent-bit-ns", "", "Fluent Bit namespace")
	flag.StringVar(&logFormat, "log-format", getEnvOrDefault("APP_LOG_FORMAT", "text"), "Log format (json or text)")
	flag.StringVar(&logLevel, "log-level", getEnvOrDefault("APP_LOG_LEVEL", "debug"), "Log level (debug, info, warn, error, fatal)")
	flag.Parse()

	if err := validateFlags(); err != nil {
		setupLog.Error(err, "Invalid flag provided")
		os.Exit(1)
	}

	ctrLogger, err := logger.New(logFormat, logLevel)
	if err != nil {
		setupLog.Error(err, "Failed to initialize logger")
		os.Exit(1)
	}
	defer func() {
		if err := ctrLogger.WithContext().Sync(); err != nil {
			setupLog.Error(err, "Failed to flush logger")
		}
	}()
	ctrl.SetLogger(zapr.NewLogger(ctrLogger.WithContext().Desugar()))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "cdd7ef0b.kyma-project.io",
	})
	if err != nil {
		setupLog.Error(err, "Failed to start manager")
		os.Exit(1)
	}

	reconciler := controllers.NewLogPipelineReconciler(
		mgr.GetClient(),
		mgr.GetScheme(),
		fluentBitNs,
		fluentBitConfigMap,
		fluentBitSectionsConfigMap,
		fluentBitParsersConfigMap,
		fluentBitDaemonSet,
		fluentBitEnvSecret,
		fluentBitFilesConfigMap)
	if err = reconciler.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "Failed to create controller", "controller", "LogPipeline")
		os.Exit(1)
	}
	if err = (&telemetryv1alpha1.LogPipeline{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "LogPipeline")
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

	setupLog.Info("Starting manager")
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
	return nil
}
