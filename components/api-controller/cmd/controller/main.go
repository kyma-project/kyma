package main

import (
	"os"
	"path/filepath"
	"time"

	istioAuthenticationClient "github.com/kyma-project/kyma/components/api-controller/pkg/clients/authentication.istio.io/clientset/versioned"
	kyma "github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma.cx/clientset/versioned"
	kymaInformers "github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma.cx/informers/externalversions"
	authenticationV2 "github.com/kyma-project/kyma/components/api-controller/pkg/controller/authentication/v2"
	"github.com/kyma-project/kyma/components/api-controller/pkg/controller/crd"
	ingressV1 "github.com/kyma-project/kyma/components/api-controller/pkg/controller/ingress/v1"
	serviceV1 "github.com/kyma-project/kyma/components/api-controller/pkg/controller/service/v1"
	"github.com/kyma-project/kyma/components/api-controller/pkg/controller/v1alpha2"
	log "github.com/sirupsen/logrus"
	apiExtensionsClient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	k8sClient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {

	log.SetLevel(getLoggerLevel())

	log.Info("Starting API controller application...")

	stop := make(chan struct{})

	kubeConfig := initKubeConfig()

	apiExtensionsClientSet := apiExtensionsClient.NewForConfigOrDie(kubeConfig)

	registerer := crd.NewRegistrar(apiExtensionsClientSet)
	registerer.Register(v1alpha2.Crd())

	k8sClientSet := k8sClient.NewForConfigOrDie(kubeConfig)

	ingressV1Interface := ingressV1.New(k8sClientSet)
	serviceV1Interface := serviceV1.New(k8sClientSet)

	istioAuthenticationClientSet := istioAuthenticationClient.NewForConfigOrDie(kubeConfig)
	authenticationV2Interface := authenticationV2.New(istioAuthenticationClientSet)

	kymaClientSet := kyma.NewForConfigOrDie(kubeConfig)

	internalInformerFactory := kymaInformers.NewSharedInformerFactory(kymaClientSet, time.Second*30)
	go internalInformerFactory.Start(stop)

	v1alpha2Controller := v1alpha2.NewController(kymaClientSet, ingressV1Interface, serviceV1Interface, authenticationV2Interface, internalInformerFactory)
	v1alpha2Controller.Run(2, stop)
}

func initKubeConfig() *rest.Config {
	kubeConfigLocation := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigLocation)
	if err != nil {
		log.Warn("unable to build kube config from file. Trying in-cluster configuration")
		kubeConfig, err = rest.InClusterConfig()
		if err != nil {
			log.Fatal("cannot find Service Account in pod to build in-cluster kube config")
		}
	}
	return kubeConfig
}

func getLoggerLevel() log.Level {

	logLevel := os.Getenv("API_CONTROLLER_LOG_LEVEL")
	if logLevel != "" {
		level, err := log.ParseLevel(logLevel)
		if err != nil {
			println("Error while setting log level: " + logLevel + ". Root cause: " + err.Error())
		} else {
			return level
		}
	}
	return log.InfoLevel
}
