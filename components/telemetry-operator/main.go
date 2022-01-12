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

	"k8s.io/apimachinery/pkg/types"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/api/v1alpha1"
	"github.com/kyma-project/kyma/components/telemetry-operator/controllers"
	//+kubebuilder:scaffold:imports
)

var (
	scheme                     = runtime.NewScheme()
	setupLog                   = ctrl.Log.WithName("setup")
	fluentBitSectionsConfigMap string
	fluentBitParsersConfigMap  string
	fluentBitDaemonSet         string
	fluentBitNs                string
	fluentBitEnvSecret         string
	fluentBitFilesConfigMap    string
)

//nolint:gochecknoinits
func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(telemetryv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&fluentBitSectionsConfigMap, "cm-name", "", "ConfigMap name to be written by Fluent Bit controller")
	flag.StringVar(&fluentBitParsersConfigMap, "parser-cm-name", "", "ConfigMap name of Fluent bit Parsers to be written by Fluent Bit controller")
	flag.StringVar(&fluentBitDaemonSet, "ds-name", "", "DaemonSet name to be managed by FluentBit controller")
	flag.StringVar(&fluentBitEnvSecret, "env-secret", "", "Secret for environment variables")
	flag.StringVar(&fluentBitFilesConfigMap, "files-cm", "", "ConfigMap for referenced files")
	flag.StringVar(&fluentBitNs, "fluent-bit-ns", "", "Fluent Bit namespace")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	if err := validateFlags(); err != nil {
		setupLog.Error(err, "invalid flag provided")
		os.Exit(1)
	}

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "cdd7ef0b.kyma-project.io",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.LogPipelineReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		FluentBitSectionsConfigMap: types.NamespacedName{
			Name:      fluentBitSectionsConfigMap,
			Namespace: fluentBitNs,
		},
		FluentBitParsersConfigMap: types.NamespacedName{
			Name:      fluentBitParsersConfigMap,
			Namespace: fluentBitNs,
		},
		FluentBitDaemonSet: types.NamespacedName{
			Name:      fluentBitDaemonSet,
			Namespace: fluentBitNs,
		},
		FluentBitEnvSecret: types.NamespacedName{
			Name:      fluentBitEnvSecret,
			Namespace: fluentBitNs,
		},
		FluentBitFilesConfigMap: types.NamespacedName{
			Name:      fluentBitFilesConfigMap,
			Namespace: fluentBitNs,
		},
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "LogPipeline")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func validateFlags() error {
	if fluentBitSectionsConfigMap == "" {
		return errors.New("--cm-name flag is required")
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

	return nil
}
