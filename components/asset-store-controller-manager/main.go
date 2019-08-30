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
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/internal/assethook"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/internal/loader"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/internal/store"
	"github.com/minio/minio-go"
	"github.com/vrischmann/envconfig"
	"net/http"
	"os"

	"github.com/kyma-project/kyma/components/asset-store-controller-manager/internal/controllers"
	assetstorev1alpha2 "github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
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

	_ = assetstorev1alpha2.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

type Config struct {
	Store         store.Config
	Loader        loader.Config
	Webhook       assethook.Config
	Asset         controllers.AssetConfig
	ClusterAsset  controllers.ClusterAssetConfig
	Bucket        controllers.BucketConfig
	ClusterBucket controllers.ClusterBucketConfig
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

	httpClient := &http.Client{}
	minioClient, err := minio.New(cfg.Store.Endpoint, cfg.Store.AccessKey, cfg.Store.SecretKey, cfg.Store.UseSSL)
	if err != nil {
		setupLog.Error(err, "unable initialize Minio client")
		os.Exit(1)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		LeaderElection:     enableLeaderElection,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	container := &controllers.Container{
		Manager:   mgr,
		Store:     store.New(minioClient, cfg.Store.UploadWorkersCount),
		Loader:    loader.New(cfg.Loader.TemporaryDirectory, cfg.Loader.VerifySSL),
		Validator: assethook.NewValidator(httpClient, cfg.Webhook.ValidationTimeout, cfg.Webhook.ValidationWorkersCount),
		Mutator:   assethook.NewMutator(httpClient, cfg.Webhook.MutationTimeout, cfg.Webhook.MutationWorkersCount),
		Extractor: assethook.NewMetadataExtractor(httpClient, cfg.Webhook.MetadataExtractionTimeout),
	}

	if err = controllers.NewClusterAsset(cfg.ClusterAsset, ctrl.Log.WithName("controllers").WithName("ClusterAsset"), container).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ClusterAsset")
		os.Exit(1)
	}
	if err = controllers.NewClusterBucket(cfg.ClusterBucket, ctrl.Log.WithName("controllers").WithName("ClusterBucket"), container).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ClusterBucket")
		os.Exit(1)
	}
	if err = controllers.NewAsset(cfg.Asset, ctrl.Log.WithName("controllers").WithName("Asset"), container).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Asset")
		os.Exit(1)
	}
	if err = controllers.NewBucket(cfg.Bucket, ctrl.Log.WithName("controllers").WithName("Bucket"), container).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Bucket")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func loadConfig(prefix string) (Config, error) {
	cfg := Config{}
	err := envconfig.InitWithPrefix(&cfg, prefix)
	if err != nil {
		return cfg, err
	}
	cfg.Bucket.ExternalEndpoint = cfg.Store.ExternalEndpoint
	cfg.ClusterBucket.ExternalEndpoint = cfg.Store.ExternalEndpoint
	return cfg, nil
}
