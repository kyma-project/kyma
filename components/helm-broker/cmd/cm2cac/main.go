package main

import (
	"os"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kyma-project/kyma/components/helm-broker/pkg/apis/addons/v1alpha1"
)

func main() {
	sch, err := v1alpha1.SchemeBuilder.Build()
	fatalOnError(err)
	err = v1.AddToScheme(sch)
	fatalOnError(err)

	namespace := os.Getenv("NAMESPACE")
	if namespace == "" {
		logrus.Fatal("Environment variable NAMESPACE must be set")
	}
	k8sConfig, err := newRestClientConfig(os.Getenv("KUBECONFIG"))
	fatalOnError(err)

	cli, err := client.New(k8sConfig, client.Options{
		Scheme: sch,
	})
	fatalOnError(err)

	logrus.Info("Start ConfigMap to ClusterAddonsConfiguration migration process")

	migrationService := NewMigrationService(cli, namespace)
	err = migrationService.Migrate()
	fatalOnError(err)

	logrus.Info("Migration done")
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
