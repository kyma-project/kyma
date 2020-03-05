/*

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
	"os"

	"github.com/kyma-project/kyma/components/function-controller/internal/controllers"
	"github.com/kyma-project/kyma/components/function-controller/pkg/apis"
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	tektonv1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	clientconfig "sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	// +kubebuilder:scaffold:imports
)

var (
	scheme             = runtime.NewScheme()
	setupLog           = ctrl.Log.WithName("setup")
	defaultMetricsAddr = ":8080"
	empty              = ""
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = serverlessv1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var appCfg configuration
	loadConfigurationOrDie(&appCfg)

	opt := zap.UseDevMode(appCfg.logDevModeEnabled)
	logger := zap.New(opt)
	ctrl.SetLogger(logger)

	setupLog.Info("Generating Kubernetes client config")
	cfg, err := clientconfig.GetConfig()
	if err != nil {
		setupLog.Error(err, "Unable to generate Kubernetes client config")
		os.Exit(1)
	}

	setupLog.Info("Initializing controller manager")
	mgr, err := manager.New(cfg, manager.Options{
		Scheme:                  scheme,
		MetricsBindAddress:      appCfg.metricsAddr,
		LeaderElection:          appCfg.leaderElectionEnabled,
		LeaderElectionID:        appCfg.leaderElectionID,
		LeaderElectionNamespace: appCfg.leaderElectionNamespace,
	})
	if err != nil {
		setupLog.Error(err, "Unable to initialize controller manager")
		os.Exit(1)
	}

	setupLog.Info("Registering custom resources...")
	schemeSetupFns := []func(*runtime.Scheme) error{
		apis.AddToScheme,
		servingv1.AddToScheme,
		tektonv1alpha1.AddToScheme,
	}

	for _, fn := range schemeSetupFns {
		if err := fn(mgr.GetScheme()); err != nil {
			setupLog.Error(err, "Unable to register custom resources")
			os.Exit(1)
		}
	}

	setupLog.Info("Custom resources registered")

	fnCtrlConf := appCfg.fnReconcilerCfg()

	ctrlCfg := &controllers.Cfg{
		Client:            mgr.GetClient(),
		Scheme:            mgr.GetScheme(),
		CacheSynchronizer: mgr.GetCache().WaitForCacheSync,
		Log:               ctrl.Log,
		EventRecorder:     mgr.GetEventRecorderFor("function-controller"),
	}

	if err := controllers.NewFunctionReconciler(ctrlCfg, fnCtrlConf).
		SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create function controller")
	}

	// TODO add other controllers here

	setupLog.Info("Running manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "Unable to run the manager")
		os.Exit(1)
	}
}
