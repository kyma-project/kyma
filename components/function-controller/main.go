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
	"flag"
	"os"

	"github.com/kelseyhightower/envconfig"
	"github.com/kyma-project/kyma/components/function-controller/pkg/container"
	configCtrl "github.com/kyma-project/kyma/components/function-controller/pkg/controllers/config"
	resource_watcher "github.com/kyma-project/kyma/components/function-controller/pkg/resource-watcher"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/dynamic"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/kyma-project/kyma/components/function-controller/pkg/apis"
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/kyma-project/kyma/components/function-controller/pkg/controllers"
	"github.com/kyma-project/kyma/components/function-controller/pkg/webhook"
	tektonv1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	// +kubebuilder:scaffold:imports
)

type Config struct {
	ResourceWatcherConfig resource_watcher.Config
}

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

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var devLog bool
	var leaderElectionCfgNamespace string
	var leaderElectionID string
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(&devLog, "devlog", false, "Enable logger's development mode")
	flag.StringVar(&leaderElectionCfgNamespace, "leader-election-configmap-namespace", "kyma-system", "The namespace in which the leader election configmap will be.")
	flag.StringVar(&leaderElectionID, "leader-election-id", "function-controller-leader-election", "Leader election ID.")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(devLog)))

	envConfig, err := loadConfig()
	failOnError(err, "unable to load config")

	setupLog.Info("Generating Kubernetes client config")
	cfg, err := config.GetConfig()
	if err != nil {
		setupLog.Error(err, "Unable to generate Kubernetes client config")
		os.Exit(1)
	}

	coreClient, err := v1.NewForConfig(cfg)
	failOnError(err, "unable to initialize core client")

	dynamicClient, err := dynamic.NewForConfig(cfg)
	failOnError(err, "unable to initialize dynamic client")

	setupLog.Info("Initializing controller manager")
	mgr, err := manager.New(cfg, manager.Options{
		Scheme:                  scheme,
		MetricsBindAddress:      metricsAddr,
		LeaderElection:          enableLeaderElection,
		LeaderElectionID:        leaderElectionID,
		LeaderElectionNamespace: leaderElectionCfgNamespace,
	})
	if err != nil {
		setupLog.Error(err, "Unable to initialize controller manager")
		os.Exit(1)
	}

	resourceWatcherServices := resource_watcher.NewResourceWatcherServices(coreClient, envConfig.ResourceWatcherConfig)
	container := &container.Container{
		Manager:                 mgr,
		CoreClient:              coreClient,
		DynamicClient:           &dynamicClient,
		ResourceWatcherServices: resourceWatcherServices,
	}

	setupLog.Info("Registering custom resources")
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

	setupLog.Info("Adding controllers to the manager")
	if err := controllers.AddToManager(mgr); err != nil {
		setupLog.Error(err, "Unable to add controllers to the manager")
		os.Exit(1)
	}

	runControllers(envConfig, container, mgr)

	setupLog.Info("setting up webhook server")
	webhook.Add(mgr)

	setupLog.Info("Running manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "Unable to run the manager")
		os.Exit(1)
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
		// Controllers for resource watcher
		"Namespace": runConfigController,
		"Secret":    runConfigController,
	}

	for name, controller := range controllers {
		err := controller(config, di, mgr, name)
		failOnError(err, "unable to create controller", "controller", name)
	}
}

func runConfigController(config Config, container *container.Container, mgr manager.Manager, name string) error {
	if !config.ResourceWatcherConfig.EnableControllers {
		return nil
	}

	return configCtrl.NewController(config.ResourceWatcherConfig, configCtrl.ResourceType(name), ctrl.Log.WithName("controllers").WithName(name), container).SetupWithManager(mgr)
}

func failOnError(err error, msg string, args ...string) {
	if err != nil {
		setupLog.Error(err, msg, args)
		os.Exit(1)
	}
}
