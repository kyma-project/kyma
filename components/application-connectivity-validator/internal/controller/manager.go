package controller

import (
	"time"

	gocache "github.com/patrickmn/go-cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	"github.com/kyma-project/kyma/components/application-connectivity-validator/internal/controller/signals"
	clientset "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	informers "github.com/kyma-project/kyma/components/application-operator/pkg/client/informers/externalversions"
)

func Start(kubeConfig string, masterURL string, syncPeriod time.Duration, appName string, cache *gocache.Cache) {
	klog.InitFlags(nil)

	stopCh := signals.SetupSignalHandler()

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeConfig)
	if err != nil {
		klog.Fatalf("Error building kubeConfig: %s", err.Error())
	}

	applicationClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building example clientset: %s", err.Error())
	}

	applicationInformerFactory := informers.NewSharedInformerFactory(applicationClient, syncPeriod)

	controller := NewController(applicationClient, applicationInformerFactory.Applicationconnector().V1alpha1().Applications(), appName, cache)
	applicationInformerFactory.Start(stopCh)

	if err = controller.Run(2, stopCh); err != nil {
		klog.Fatalf("Error running controller: %s", err.Error())
	}
}
