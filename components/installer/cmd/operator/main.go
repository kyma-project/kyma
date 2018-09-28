package main

import (
	"flag"
	"log"
	"time"

	"github.com/kyma-project/kyma/components/installer/pkg/actionmanager"
	"github.com/kyma-project/kyma/components/installer/pkg/conditionmanager"
	"github.com/kyma-project/kyma/components/installer/pkg/consts"
	"github.com/kyma-project/kyma/components/installer/pkg/finalizer"
	"github.com/kyma-project/kyma/components/installer/pkg/installation"
	"github.com/kyma-project/kyma/components/installer/pkg/kymahelm"
	"github.com/kyma-project/kyma/components/installer/pkg/kymasources"
	"github.com/kyma-project/kyma/components/installer/pkg/servicecatalog"
	"github.com/kyma-project/kyma/components/installer/pkg/toolkit"

	"github.com/kyma-project/kyma/components/installer/pkg/statusmanager"

	"github.com/kyma-project/kyma/components/installer/pkg/steps"

	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	clientset "github.com/kyma-project/kyma/components/installer/pkg/client/clientset/versioned"
	informers "github.com/kyma-project/kyma/components/installer/pkg/client/informers/externalversions"
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

	helmClient := kymahelm.NewClient(*helmHost)
	serviceCatalogClient := servicecatalog.NewClient(config)
	kymaCommandExecutor := &toolkit.KymaCommandExecutor{}

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
	internalInformerFactory := informers.NewSharedInformerFactory(internalClient, time.Second*30)
	installationLister := internalInformerFactory.Installer().V1alpha1().Installations().Lister()

	kymaStatusManager := statusmanager.NewKymaStatusManager(internalClient, installationLister)
	kymaActionManager := actionmanager.NewKymaActionManager(internalClient, installationLister)
	conditionManager := conditionmanager.New(internalClient, installationLister)

	installationFinalizerManager := finalizer.NewManager(consts.InstFinalizer)

	kymaPackages := kymasources.NewKymaPackages(kymasources.NewFilesystemWrapper(), kymaCommandExecutor, *kymaDir)

	installationSteps := steps.New(helmClient, kubeClient, serviceCatalogClient, kymaStatusManager, kymaActionManager, kymaCommandExecutor, kymaPackages)

	installationController := installation.NewController(kubeClient, kubeInformerFactory, internalInformerFactory, installationSteps, conditionManager, installationFinalizerManager, internalClient)

	kubeInformerFactory.Start(stop)
	internalInformerFactory.Start(stop)

	installationController.Run(2, stop)
}

func getClientConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}
