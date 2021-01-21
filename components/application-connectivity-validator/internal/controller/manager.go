package controller

import (
	"time"

	"github.com/kyma-project/kyma/common/logger/logger"

	"github.com/kyma-project/kyma/components/application-connectivity-validator/internal/controller/signals"
	clientset "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	informers "github.com/kyma-project/kyma/components/application-operator/pkg/client/informers/externalversions"
	gocache "github.com/patrickmn/go-cache"
	"k8s.io/client-go/tools/clientcmd"
)

func Start(log *logger.Logger, kubeConfig string, masterURL string, syncPeriod time.Duration, appName string, cache *gocache.Cache) {
	stopCh := signals.SetupSignalHandler()

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeConfig)
	if err != nil {
		log.WithContext().With("masterURL", masterURL).With("kubeConfig", kubeConfig).Fatalf("Failed to build kubeConfig: %s", err.Error())
	}

	applicationClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		log.WithContext().Fatalf("Failed to build application clientset: %s", err.Error())
	}

	applicationInformerFactory := informers.NewSharedInformerFactory(applicationClient, syncPeriod)

	controller := NewController(log, applicationClient, applicationInformerFactory.Applicationconnector().V1alpha1().Applications(), appName, cache)
	applicationInformerFactory.Start(stopCh)

	if err = controller.Run(2, stopCh); err != nil {
		log.WithContext().With("applicationName", appName).With("controller", controllerName).Fatalf("While running controller: %s", err.Error())
	}
}
