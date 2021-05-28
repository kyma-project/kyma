package controller

import (
	"time"

	"github.com/kyma-project/kyma/common/logging/logger"
	clientset "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	informers "github.com/kyma-project/kyma/components/application-operator/pkg/client/informers/externalversions"
	"github.com/kyma-project/kyma/components/central-application-connectivity-validator/internal/controller/signals"
	gocache "github.com/patrickmn/go-cache"
	"k8s.io/client-go/tools/clientcmd"
)

func Start(log *logger.Logger, kubeConfig string, apiServerURL string, syncPeriod time.Duration, cache *gocache.Cache) {
	stopCh := signals.SetupSignalHandler()

	cfg, err := clientcmd.BuildConfigFromFlags(apiServerURL, kubeConfig)
	if err != nil {
		log.WithContext().With("apiServerURL", apiServerURL).With("kubeConfig", kubeConfig).Fatalf("Failed to build kubeConfig: %s", err.Error())
	}

	applicationClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		log.WithContext().Fatalf("Failed to build application clientset: %s", err.Error())
	}

	applicationInformerFactory := informers.NewSharedInformerFactory(applicationClient, syncPeriod)

	controller := NewController(log, applicationInformerFactory.Applicationconnector().V1alpha1().Applications(), cache)
	applicationInformerFactory.Start(stopCh)

	if err = controller.Run(2, stopCh); err != nil {
		log.WithContext().With("controller", controllerName).Fatalf("While running controller: %s", err.Error())
	}
}
