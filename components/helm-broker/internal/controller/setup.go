package controller

import (
	"github.com/kyma-project/kyma/components/helm-broker/internal/controller/broker"
	"github.com/sirupsen/logrus"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	"github.com/kyma-project/kyma/components/helm-broker/internal/bundle"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	"github.com/kyma-project/kyma/components/helm-broker/internal/config"
	"github.com/kyma-project/kyma/components/helm-broker/internal/storage"
	"github.com/kyma-project/kyma/components/helm-broker/pkg/apis"
)

func SetupAndStartController(cfg *rest.Config, ctrCfg *config.ControllerConfig, metricsAddr string, sFact storage.Factory, log *logrus.Entry) manager.Manager {
	// Create a new Cmd to provide shared dependencies and start components
	log.Info("Setting up manager")
	mgr, err := manager.New(cfg, manager.Options{MetricsBindAddress: metricsAddr})
	fatalOnError(err, "while setting up a manager")

	log.Info("Registering Components.")

	// Setup Scheme for all resources
	log.Info("Setting up schemes")
	err = apis.AddToScheme(mgr.GetScheme())
	fatalOnError(err, "while adding AC scheme")
	err = v1beta1.AddToScheme(mgr.GetScheme())
	fatalOnError(err, "while adding SC scheme")
	err = v1alpha1.AddToScheme(mgr.GetScheme())
	fatalOnError(err, "while adding CMS scheme")

	docsProvider := NewDocsProvider(mgr.GetClient())
	brokerSyncer := broker.NewServiceBrokerSyncer(mgr.GetClient(), ctrCfg.ClusterServiceBrokerName, log)
	sbFacade := broker.NewBrokersFacade(mgr.GetClient(), brokerSyncer, ctrCfg.Namespace, ctrCfg.ServiceName, log)
	csbFacade := broker.NewClusterBrokersFacade(mgr.GetClient(), brokerSyncer, ctrCfg.Namespace, ctrCfg.ServiceName, ctrCfg.ClusterServiceBrokerName, log)
	bundleProvider := bundle.NewProvider(bundle.NewHTTPRepository(), bundle.NewLoader(ctrCfg.TmpDir, log), log)

	log.Info("Setting up controller")
	acReconcile := NewReconcileAddonsConfiguration(mgr, bundleProvider, sFact.Chart(), sFact.Bundle(), sbFacade, docsProvider, brokerSyncer, ctrCfg.DevelopMode)
	acController := NewAddonsConfigurationController(acReconcile)
	err = acController.Start(mgr)
	fatalOnError(err, "unable to start AddonsConfigurationController")

	cacReconcile := NewReconcileClusterAddonsConfiguration(mgr, bundleProvider, sFact.Chart(), sFact.Bundle(), csbFacade, docsProvider, brokerSyncer, ctrCfg.DevelopMode)
	cacController := NewClusterAddonsConfigurationController(cacReconcile)
	err = cacController.Start(mgr)

	if err != nil {
		log.Error(err, "unable to start ClusterAddonsConfigurationController")
	}
	return mgr
}

func fatalOnError(err error, msg string) {
	if err != nil {
		logrus.Fatalf("%s: %s", msg, err.Error())
	}
}