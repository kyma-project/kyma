package controller

import (
	"go.uber.org/zap"
	"time"

	"github.com/kyma-project/kyma/components/application-connectivity-validator/internal/controller/signals"
	clientset "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	informers "github.com/kyma-project/kyma/components/application-operator/pkg/client/informers/externalversions"
	gocache "github.com/patrickmn/go-cache"
	"k8s.io/client-go/tools/clientcmd"
)

func Start(log *zap.SugaredLogger, kubeConfig string, masterURL string, syncPeriod time.Duration, appName string, cache *gocache.Cache) {
	stopCh := signals.SetupSignalHandler()
	//TODO: setup the logger for k8s controllers
	// https://www.bookstack.cn/read/Operator-SDK-1.0-en/f8568aba96fc669e.md
	// https://github.com/kubernetes-sigs/controller-runtime/blob/master/pkg/log/zap/zap.go
	// client-go logs with hardcoded klog - https://github.com/kubernetes/kubernetes/issues/94428

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeConfig)
	if err != nil {
		log.Fatalf("Error building kubeConfig: %s", err.Error())
	}

	applicationClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("Error building example clientset: %s", err.Error())
	}

	applicationInformerFactory := informers.NewSharedInformerFactory(applicationClient, syncPeriod)

	controller := NewController(log, applicationClient, applicationInformerFactory.Applicationconnector().V1alpha1().Applications(), appName, cache)
	applicationInformerFactory.Start(stopCh)

	if err = controller.Run(2, stopCh); err != nil {
		log.Fatalf("Error running controller: %s", err.Error())
	}
}
