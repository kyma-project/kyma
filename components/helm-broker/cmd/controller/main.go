package main

import (
	"flag"

	"github.com/kyma-project/kyma/components/helm-broker/internal/controller"
	"github.com/kyma-project/kyma/components/helm-broker/internal/storage"
	"github.com/kyma-project/kyma/components/helm-broker/pkg/apis"

	envs "github.com/kyma-project/kyma/components/helm-broker/internal/config"
	"github.com/kyma-project/kyma/components/helm-broker/internal/controller/broker"
	"github.com/kyma-project/kyma/components/helm-broker/platform/logger"
	"github.com/sirupsen/logrus"

	"fmt"
	"time"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	"github.com/kyma-project/kyma/components/helm-broker/internal/bundle"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

func main() {
	var metricsAddr string
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	verbose := flag.Bool("verbose", false, "specify if lg verbosely loading configuration")
	flag.Parse()

	ctrCfg, err := envs.LoadControllerConfig(*verbose)
	fatalOnError(err, "while loading config")

	storageConfig := storage.ConfigList(ctrCfg.Storage)
	sFact, err := storage.NewFactory(&storageConfig)
	fatalOnError(err, "while setting up a storage")

	lg := logger.New(&ctrCfg.Logger)

	// Get a config to talk to the apiserver
	lg.Info("Setting up client for manager")
	cfg, err := config.GetConfig()
	fatalOnError(err, "while setting up a client")

	// Create a new Cmd to provide shared dependencies and start components
	lg.Info("Setting up manager")
	var mgr manager.Manager
	fatalOnError(waitAtMost(func() (bool, error) {
		mgr, err = manager.New(cfg, manager.Options{MetricsBindAddress: metricsAddr})
		if err != nil {
			return false, err
		}
		return true, nil
	}, time.Minute), "while setting up a manager")

	lg.Info("Registering Components.")

	// Setup Scheme for all resources
	lg.Info("Setting up schemes")
	fatalOnError(apis.AddToScheme(mgr.GetScheme()), "while adding AC scheme")
	fatalOnError(v1beta1.AddToScheme(mgr.GetScheme()), "while adding SC scheme")
	fatalOnError(v1alpha1.AddToScheme(mgr.GetScheme()), "while adding CMS scheme")

	docsProvider := controller.NewDocsProvider(mgr.GetClient())
	brokerSyncer := broker.NewServiceBrokerSyncer(mgr.GetClient(), ctrCfg.ClusterServiceBrokerName, lg)
	sbFacade := broker.NewBrokersFacade(mgr.GetClient(), brokerSyncer, ctrCfg.Namespace, ctrCfg.ServiceName, lg)
	csbFacade := broker.NewClusterBrokersFacade(mgr.GetClient(), brokerSyncer, ctrCfg.Namespace, ctrCfg.ServiceName, ctrCfg.ClusterServiceBrokerName, lg)
	bundleProvider := bundle.NewProvider(bundle.NewHTTPRepository(), bundle.NewLoader(ctrCfg.TmpDir, lg), lg)

	lg.Info("Setting up controller")
	acReconcile := controller.NewReconcileAddonsConfiguration(mgr, bundleProvider, sFact.Chart(), sFact.Bundle(), sbFacade, docsProvider, brokerSyncer, ctrCfg.DevelopMode)
	acController := controller.NewAddonsConfigurationController(acReconcile)
	err = acController.Start(mgr)
	fatalOnError(err, "unable to start AddonsConfigurationController")

	cacReconcile := controller.NewReconcileClusterAddonsConfiguration(mgr, bundleProvider, sFact.Chart(), sFact.Bundle(), csbFacade, docsProvider, brokerSyncer, ctrCfg.DevelopMode)
	cacController := controller.NewClusterAddonsConfigurationController(cacReconcile)
	err = cacController.Start(mgr)
	fatalOnError(err, "unable to start ClusterAddonsConfigurationController")

	lg.Info("Starting the Controller.")
	err = mgr.Start(signals.SetupSignalHandler())
	fatalOnError(err, "unable to run the manager")
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
