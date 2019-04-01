package main

import (
	"time"

	"github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	"k8s.io/client-go/rest"

	"github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned/scheme"
	"github.com/kyma-project/kyma/components/application-operator/pkg/controller"
	"github.com/kyma-project/kyma/components/application-operator/pkg/kymahelm"
	appRelease "github.com/kyma-project/kyma/components/application-operator/pkg/kymahelm/application"
	log "github.com/sirupsen/logrus"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

func main() {
	formatter := &log.TextFormatter{
		FullTimestamp: true,
	}
	log.SetFormatter(formatter)

	log.Info("Starting Application Operator.")

	options := parseArgs()
	log.Infof("Options: %s", options)

	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatal(err)
	}

	syncPeriod := time.Second * time.Duration(options.syncPeriod)
	mgrOpts := manager.Options{
		SyncPeriod: &syncPeriod,
	}

	mgr, err := manager.New(cfg, mgrOpts)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Registering Components:")

	log.Printf("Setting up scheme.")

	scheme.AddToScheme(mgr.GetScheme())

	log.Printf("Preparing Release Manager.")

	releaseManager, err := newReleaseManager(options, cfg)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Upgrading releases.")

	err = releaseManager.UpgradeReleases()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Setting up controller.")

	err = controller.InitApplicationController(mgr, releaseManager, options.appName)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Starting the Cmd.")
	log.Info(mgr.Start(signals.SetupSignalHandler()))
}

func newReleaseManager(options *options, cfg *rest.Config) (appRelease.ReleaseManager, error) {
	overridesDefaults := appRelease.OverridesData{
		DomainName:                   options.domainName,
		ApplicationGatewayImage:      options.applicationGatewayImage,
		ApplicationGatewayTestsImage: options.applicationGatewayTestsImage,
		EventServiceImage:            options.eventServiceImage,
		EventServiceTestsImage:       options.eventServiceTestsImage,
	}

	appClient, err := versioned.NewForConfig(cfg)
	if err != nil {
		log.Fatal(err)
	}

	helmClient, err := kymahelm.NewClient(options.tillerUrl, options.helmTLSKeyFile, options.helmTLSCertificateFile, options.tillerTLSSkipVerify, options.installationTimeout)
	if err != nil {
		return nil, err
	}

	releaseManager := appRelease.NewReleaseManager(helmClient, appClient.ApplicationconnectorV1alpha1().Applications(), overridesDefaults, options.namespace)

	return releaseManager, nil
}
