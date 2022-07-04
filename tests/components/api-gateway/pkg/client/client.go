package client

import (
	"os"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const kubeconfigEnvName = "KUBECONFIG"

func loadKubeConfigOrDie() *rest.Config {
	if kubeconfig, ok := os.LookupEnv(kubeconfigEnvName); ok {
		cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			panic(err)
		}
		return cfg
	}

	if _, err := os.Stat(clientcmd.RecommendedHomeFile); os.IsNotExist(err) {
		cfg, err := rest.InClusterConfig()
		if err != nil {
			panic(err)
		}
		return cfg
	}

	cfg, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		panic(err)
	}
	return cfg
}

func GetDynamicClient() dynamic.Interface {
	client, err := dynamic.NewForConfig(loadKubeConfigOrDie())
	if err != nil {
		panic(err)
	}
	return client
}
