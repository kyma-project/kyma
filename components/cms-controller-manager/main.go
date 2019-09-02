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
	"time"

	assetstore "github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/components/cms-controller-manager/internal/controllers"
	"github.com/kyma-project/kyma/components/cms-controller-manager/internal/webhookconfig"
	cmsv1alpha1 "github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	// External
	_ = assetstore.AddToScheme(scheme)

	_ = cmsv1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

type Config struct {
	DocsTopic           controllers.DocsTopicConfig
	ClusterDocsTopic    controllers.ClusterDocsTopicConfig
	Webhook             webhookconfig.Config
	BucketRegion        string `envconfig:"optional"`
	ClusterBucketRegion string `envconfig:"optional"`
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	ctrl.SetLogger(zap.Logger(true))

	cfg, err := loadConfig("APP")
	if err != nil {
		setupLog.Error(err, "unable to load config")
		os.Exit(1)
	}

	restConfig := ctrl.GetConfigOrDie()
	mgr, err := ctrl.NewManager(restConfig, ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "cms-controller-leader-election-helper",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	webhookSvc, cfgMapInformer, err := initWebhookConfigService(cfg.Webhook, restConfig)
	if err != nil {
		setupLog.Error(err, "unable to initialize webhook service")
		os.Exit(1)
	}

	if err = controllers.NewClusterDocsTopic(cfg.ClusterDocsTopic, ctrl.Log.WithName("controllers").WithName("ClusterDocsTopic"), mgr, webhookSvc).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ClusterDocsTopic")
		os.Exit(1)
	}
	if err = controllers.NewDocsTopic(cfg.DocsTopic, ctrl.Log.WithName("controllers").WithName("DocsTopic"), mgr, webhookSvc).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "DocsTopic")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	setupSignal := ctrl.SetupSignalHandler()
	go cfgMapInformer.Run(setupSignal)

	setupLog.Info("starting manager")
	if err := mgr.Start(setupSignal); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func loadConfig(prefix string) (Config, error) {
	cfg := Config{}
	err := envconfig.InitWithPrefix(&cfg, prefix)
	if err != nil {
		return Config{}, err
	}
	cfg.ClusterDocsTopic.BucketRegion = cfg.ClusterBucketRegion
	cfg.DocsTopic.BucketRegion = cfg.BucketRegion
	return cfg, err
}

func initWebhookConfigService(webhookCfg webhookconfig.Config, config *rest.Config) (webhookconfig.AssetWebhookConfigService, cache.SharedIndexInformer, error) {
	dc, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, nil, errors.Wrap(err, "while initializing dynamic client")
	}

	informerFactory := dynamicinformer.NewDynamicSharedInformerFactory(dc, time.Hour)
	configmapsResource := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}
	informer := informerFactory.ForResource(configmapsResource).Informer()

	webhookCfgService := webhookconfig.New(informer.GetIndexer(), webhookCfg.CfgMapName, webhookCfg.CfgMapNamespace)
	return webhookCfgService, informer, nil
}
