package main

import (
	"flag"
	"log"
	"time"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/actionmanager"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/conditionmanager"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/consts"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/finalizer"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/installation"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/kymahelm"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/kymainstallation"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/kymasources"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/servicecatalog"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/toolkit"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/statusmanager"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/steps"

	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	clientset "github.com/kyma-project/kyma/components/kyma-operator/pkg/client/clientset/versioned"
	informers "github.com/kyma-project/kyma/components/kyma-operator/pkg/client/informers/externalversions"
)

var gitCommitHash string

func main() {

	log.Println("Starting operator...")

	if gitCommitHash != "" {
		log.Println("Git commit hash:", gitCommitHash)
	}

	stop := make(chan struct{})

	kubeconfig := flag.String("kubeconfig", "", "Path to a kubeconfig file")
	helmHost := flag.String("helmhost", "tiller-deploy.kube-system.svc.cluster.local:44134", "Helm host")
	kymaDir := flag.String("kymadir", "/kyma", "Directory where kyma packages will be extracted")
	tlsKey := flag.String("tillerTLSKey", "/etc/certs/tls.key", "Path to TLS key file")
	tlsCrt := flag.String("tillerTLSCrt", "/etc/certs/tls.crt", "Path to TLS cert file")
	TLSInsecureSkipVerify := flag.Bool("tillerTLSInsecureSkipVerify", false, "Disable verification of Tiller TLS cert")
	backoffIntervals := flag.String("retryIntervals", "[10, 20, 30, 40, 50]", "TODO")

	flag.Parse()

	config, err := getClientConfig(*kubeconfig)

	if err != nil {
		log.Fatalf("Unable to build kubernetes configuration. Error: %v", err)
	}

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Unable to create kubernetes client. Error: %v", err)
	}

	internalClient, err := clientset.NewForConfig(config)
	if err != nil {
		log.Fatalf("Unable to create internal client. Error: %v", err)
	}

	helmClient, err := kymahelm.NewClient(*helmHost, *tlsKey, *tlsCrt, *TLSInsecureSkipVerify)
	if err != nil {
		log.Fatalf("Unable create helm client. Error: %v", err)
	}

	serviceCatalogClient := servicecatalog.NewClient(config)
	kymaCommandExecutor := &toolkit.KymaCommandExecutor{}

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
	internalInformerFactory := informers.NewSharedInformerFactory(internalClient, time.Second*30)
	installationLister := internalInformerFactory.Installer().V1alpha1().Installations().Lister()

	kymaStatusManager := statusmanager.NewKymaStatusManager(internalClient, installationLister)
	kymaActionManager := actionmanager.NewKymaActionManager(internalClient, installationLister)
	conditionManager := conditionmanager.New(internalClient, installationLister)

	installationFinalizerManager := finalizer.NewManager(consts.InstFinalizer)

	fsWrapper := kymasources.NewFilesystemWrapper()

	kymaPackages := kymasources.NewKymaPackages(fsWrapper, kymaCommandExecutor, *kymaDir)
	stepFactoryCreator := kymainstallation.NewStepFactoryCreator(helmClient, kymaPackages, fsWrapper, *kymaDir)
	installationSteps := steps.New(serviceCatalogClient, kymaStatusManager, kymaActionManager, stepFactoryCreator, parseBackoffIntervals(*backoffIntervals))

	installationController := installation.NewController(kubeClient, kubeInformerFactory, internalInformerFactory, installationSteps, conditionManager, installationFinalizerManager, internalClient)

	kubeInformerFactory.Start(stop)
	internalInformerFactory.Start(stop)

	installationController.Run(stop)
}

func getClientConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}

func parseBackoffIntervals(backoffIntervals string) []uint {
	backoffIntervalsParsed := []uint{10, 20, 30, 40, 50} // TODO
	return backoffIntervalsParsed
}
