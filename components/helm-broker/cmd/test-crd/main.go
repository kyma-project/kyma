package main

import (
	"github.com/kyma-project/kyma/components/helm-broker/pkg/apis/addons/v1alpha1"
	addonsClientset "github.com/kyma-project/kyma/components/helm-broker/pkg/client/clientset/versioned/typed/addons/v1alpha1"
	"github.com/sirupsen/logrus"
	"github.com/vrischmann/envconfig"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Config holds configuration for the sample test CRD
type Config struct {
	Kubeconfig string
}

// This application is only for showing that generated clients can perform the CRUD operations
// To test that you can execute:
// 		env KUBECONFIG=<path> go run cmd/test-crd/main.go
func main() {
	var cfg Config

	fatalOnError(envconfig.Init(&cfg))
	// creates the in-cluster k8sConfig
	k8sConfig, err := newRestClientConfig(cfg.Kubeconfig)
	fatalOnError(err)

	addonsCli, err := addonsClientset.NewForConfig(k8sConfig)
	fatalOnError(err)

	// AddonsConfigurations CRUD
	addonCfg, err := addonsCli.AddonsConfigurations("default").Create(&v1alpha1.AddonsConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: "testing-configuration",
		},
		Spec: v1alpha1.AddonsConfigurationSpec{
			CommonAddonsConfigurationSpec: v1alpha1.CommonAddonsConfigurationSpec{
				Repositories: []v1alpha1.SpecRepository{
					{URL: "https://github.com/kyma-project/bundles/releases/download/0.6.0/index.yaml"},
				},
			},
		},
	})
	fatalOnError(err)

	addonCfg, err = addonsCli.AddonsConfigurations(addonCfg.Namespace).Get(addonCfg.Name, metav1.GetOptions{})
	fatalOnError(err)

	err = addonsCli.AddonsConfigurations(addonCfg.Namespace).Delete(addonCfg.Name, &metav1.DeleteOptions{})
	fatalOnError(err)

	// ClusterAddonsConfigurations CRUD
	clusterAddonCfg, err := addonsCli.ClusterAddonsConfigurations().Create(&v1alpha1.ClusterAddonsConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: "testing-configuration",
		},
		Spec: v1alpha1.ClusterAddonsConfigurationSpec{
			CommonAddonsConfigurationSpec: v1alpha1.CommonAddonsConfigurationSpec{
				Repositories: []v1alpha1.SpecRepository{
					{URL: "https://github.com/kyma-project/bundles/releases/download/0.6.0/index.yaml"},
				},
			},
		},
	})
	fatalOnError(err)

	clusterAddonCfg, err = addonsCli.ClusterAddonsConfigurations().Get(clusterAddonCfg.Name, metav1.GetOptions{})
	fatalOnError(err)

	err = addonsCli.ClusterAddonsConfigurations().Delete(clusterAddonCfg.Name, &metav1.DeleteOptions{})
	fatalOnError(err)
}

func newRestClientConfig(kubeConfigPath string) (*rest.Config, error) {
	if kubeConfigPath != "" {
		return clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	}

	return rest.InClusterConfig()
}

func fatalOnError(err error) {
	if err != nil {
		logrus.Fatal(err.Error())
	}
}
