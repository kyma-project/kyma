package main

import (
	"os"
	"path/filepath"
	"time"

	istioAuthenticationClient "github.com/kyma-project/kyma/components/api-controller/pkg/clients/authentication.istio.io/clientset/versioned"
	kyma "github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma-project.io/clientset/versioned"
	kymaInformers "github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma-project.io/informers/externalversions"
	istioNetworkingClient "github.com/kyma-project/kyma/components/api-controller/pkg/clients/networking.istio.io/clientset/versioned"
	authenticationV2 "github.com/kyma-project/kyma/components/api-controller/pkg/controller/authentication/v2"
	"github.com/kyma-project/kyma/components/api-controller/pkg/controller/crd"
	istioNetworkingV1 "github.com/kyma-project/kyma/components/api-controller/pkg/controller/networking/v1"
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

	jwtDefaultConfig := initJwtDefaultConfig()

	mTLSOptionEnabled := isAuthPolicyMTLSEnabled()

	istioGateway := getIstioGateway()

	kubeConfig := initKubeConfig()

	domainName := initDomainName()

	apiExtensionsClientSet := apiExtensionsClient.NewForConfigOrDie(kubeConfig)

	registerer := crd.NewRegistrar(apiExtensionsClientSet)
	registerer.Register(v1alpha2.Crd(domainName))

	k8sClientSet := k8sClient.NewForConfigOrDie(kubeConfig)
	serviceV1Interface := serviceV1.New(k8sClientSet)

	istioNetworkingClientSet := istioNetworkingClient.NewForConfigOrDie(kubeConfig)
	istioNetworkingV1Interface := istioNetworkingV1.New(istioNetworkingClientSet, k8sClientSet, istioGateway)

	istioAuthenticationClientSet := istioAuthenticationClient.NewForConfigOrDie(kubeConfig)
	authenticationV2Interface := authenticationV2.New(istioAuthenticationClientSet, jwtDefaultConfig, mTLSOptionEnabled)

	kymaClientSet := kyma.NewForConfigOrDie(kubeConfig)

	internalInformerFactory := kymaInformers.NewSharedInformerFactory(kymaClientSet, time.Second*30)

	v1alpha2Controller := v1alpha2.NewController(kymaClientSet, istioNetworkingV1Interface, serviceV1Interface, authenticationV2Interface, internalInformerFactory, domainName)
	internalInformerFactory.Start(stop)
	err := v1alpha2Controller.Run(2, stop)
	if err != nil {
		log.Fatalf("Unable to run controller: %v", err)
	}
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

func getIstioGateway() string {
	gateway := os.Getenv("GATEWAY_FQDN")

	if gateway == "" {
		log.Fatal("gateway not provided. Please provide env variables GATEWAY_FQDN")
	}
	return gateway
}

func initJwtDefaultConfig() authenticationV2.JwtDefaultConfig {
	issuer := os.Getenv("DEFAULT_ISSUER")
	jwksURI := os.Getenv("DEFAULT_JWKS_URI")

	if issuer == "" || jwksURI == "" {
		log.Fatal("default issuer or jwksURI not provided. Please provide env variables DEFAULT_ISSUER and DEFAULT_JWKS_URI")
	}
	return authenticationV2.JwtDefaultConfig{
		Issuer:  issuer,
		JwksUri: jwksURI,
	}
}

func initDomainName() string {
	domainName := os.Getenv("DOMAIN_NAME")

	if domainName == "" {
		log.Fatal("domain name not provided. Please provide env variable DOMAIN_NAME")
	}

	return domainName
}

func isAuthPolicyMTLSEnabled() bool {
	return os.Getenv("ENABLE_MTLS") == "true"
}
