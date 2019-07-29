package controller

import (
	"fmt"
	"time"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	"github.com/kyma-project/kyma/components/helm-broker/internal/addon"
	"github.com/kyma-project/kyma/components/helm-broker/internal/config"
	"github.com/kyma-project/kyma/components/helm-broker/internal/controller/broker"
	"github.com/kyma-project/kyma/components/helm-broker/internal/storage"
	"github.com/kyma-project/kyma/components/helm-broker/pkg/apis"
)

// SetupAndStartController creates and starts the controller
func SetupAndStartController(cfg *rest.Config, ctrCfg *config.ControllerConfig, metricsAddr string, sFact storage.Factory, lg *logrus.Entry) manager.Manager {
	// Create a new Cmd to provide shared dependencies and start components
	lg.Info("Setting up manager")
	var mgr manager.Manager
	fatalOnError(waitAtMost(func() (bool, error) {
		newMgr, err := manager.New(cfg, manager.Options{MetricsBindAddress: metricsAddr})
		if err != nil {
			return false, err
		}
		mgr = newMgr
		return true, nil
	}, time.Minute), "while setting up a manager")

	lg.Info("Registering Components.")

	// Setup Scheme for all resources
	lg.Info("Setting up schemes")
	fatalOnError(apis.AddToScheme(mgr.GetScheme()), "while adding AC scheme")
	fatalOnError(v1beta1.AddToScheme(mgr.GetScheme()), "while adding SC scheme")
	fatalOnError(v1alpha1.AddToScheme(mgr.GetScheme()), "while adding CMS scheme")

	docsProvider := NewDocsProvider(mgr.GetClient())
	brokerSyncer := broker.NewServiceBrokerSyncer(mgr.GetClient(), ctrCfg.ClusterServiceBrokerName, lg)
	sbFacade := broker.NewBrokersFacade(mgr.GetClient(), brokerSyncer, ctrCfg.Namespace, ctrCfg.ServiceName, lg)
	csbFacade := broker.NewClusterBrokersFacade(mgr.GetClient(), brokerSyncer, ctrCfg.Namespace, ctrCfg.ServiceName, ctrCfg.ClusterServiceBrokerName, lg)
	addonProvider := addon.NewProvider(addon.NewHTTPRepository(), addon.NewLoader(ctrCfg.TmpDir, lg), lg)

	lg.Info("Setting up controller")
	acReconcile := NewReconcileAddonsConfiguration(mgr, addonProvider, sFact.Chart(), sFact.Addon(), sbFacade, docsProvider, brokerSyncer, ctrCfg.DevelopMode)
	acController := NewAddonsConfigurationController(acReconcile)
	err := acController.Start(mgr)
	fatalOnError(err, "unable to start AddonsConfigurationController")

	cacReconcile := NewReconcileClusterAddonsConfiguration(mgr, addonProvider, sFact.Chart(), sFact.Addon(), csbFacade, docsProvider, brokerSyncer, ctrCfg.DevelopMode)
	cacController := NewClusterAddonsConfigurationController(cacReconcile)
	err = cacController.Start(mgr)
	fatalOnError(err, "unable to start ClusterAddonsConfigurationController")

	return mgr
}

func fatalOnError(err error, msg string) {
	if err != nil {
		logrus.Fatalf("%s: %s", msg, err.Error())
	}
}

func waitAtMost(fn func() (bool, error), duration time.Duration) error {
	timeout := time.After(duration)
	tick := time.Tick(500 * time.Millisecond)

	for {
		ok, err := fn()
		select {
		case <-timeout:
			return fmt.Errorf("waiting for resource failed in given timeout %f second(s)", duration.Seconds())
		case <-tick:
			if err != nil {
				logrus.Println(err)
			} else if ok {
				return nil
			}
		}
	}
}
