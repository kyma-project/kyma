package main

import (
	"flag"
	"log"

	"github.com/kyma-project/kyma/components/environments/internal/controller"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {

	kubeconfig := flag.String("kubeconfig", "", "Path to a kubeconfig file")

	var cfg controller.EnvironmentsConfig
	err := envconfig.InitWithPrefix(&cfg, "APP")
	panicOnError(errors.Wrap(err, "while reading configuration from environment variables"))

	flag.Parse()

	config, err := getClientConfig(*kubeconfig)
	panicOnError(err)

	clientset, err := kubernetes.NewForConfig(config)
	panicOnError(err)

	controllerInstance, err := controller.NewController(clientset, &cfg)
	panicOnError(err)

	stop := make(chan struct{})
	go controllerInstance.Run(stop)

	log.Println("Started listening")

	<-stop
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func getClientConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}
