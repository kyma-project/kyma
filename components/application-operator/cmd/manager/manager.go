package main

import (
	"time"

	service_instance_scheme "github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset/typed/servicecatalog/v1beta1"

	"github.com/kyma-project/kyma/components/application-operator/pkg/kymahelm/gateway"

	service_instance_controller "github.com/kyma-project/kyma/components/application-operator/pkg/service-instance-controller"

	"github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	"k8s.io/client-go/rest"

	application_controller "github.com/kyma-project/kyma/components/application-operator/pkg/application-controller"
	"github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned/scheme"
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

	err = scheme.AddToScheme(mgr.GetScheme())
	if err != nil {
		log.Fatal(err)
	}

	err = service_instance_scheme.AddToScheme(mgr.GetScheme())
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Preparing Helm Client.")

	helmClient, err := kymahelm.NewClient(options.tillerUrl, options.helmTLSKeyFile, options.helmTLSCertificateFile, options.tillerTLSSkipVerify, options.installationTimeout)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Preparing Release Manager.")

	releaseManager, err := newApplicationReleaseManager(options, cfg, helmClient)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Upgrading releases.")

	err = releaseManager.UpgradeApplicationReleases()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Preparing Gateway Manager.")

	gatewayManager, err := newGatewayManager(options, cfg, helmClient)

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Upgrading gateways")

	err = gatewayManager.UpgradeGateways()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Setting up Application Controller.")

	err = application_controller.InitApplicationController(mgr, releaseManager, options.appName)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Setting up Service Instance Controller.")

	err = service_instance_controller.InitServiceInstanceController(mgr, options.appName, gatewayManager)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Starting the Cmd.")
	log.Info(mgr.Start(signals.SetupSignalHandler()))
}

func newGatewayManager(options *options, cfg *rest.Config, helmClient kymahelm.HelmClient) (gateway.GatewayManager, error) {
	overrides := gateway.OverridesData{
		ApplicationGatewayImage:      options.applicationGatewayImage,
		ApplicationGatewayTestsImage: options.applicationGatewayTestsImage,
	}

	serviceCatalogueClient, err := v1beta1.NewForConfig(cfg)

	if err != nil {
		return nil, err
	}

	return gateway.NewGatewayManager(helmClient, overrides, serviceCatalogueClient.ServiceInstances("")), nil
}

func newApplicationReleaseManager(options *options, cfg *rest.Config, helmClient kymahelm.HelmClient) (appRelease.ApplicationReleaseManager, error) {
	overridesDefaults := appRelease.OverridesData{
		DomainName:                            options.domainName,
		ApplicationGatewayImage:               options.applicationGatewayImage,
		ApplicationGatewayTestsImage:          options.applicationGatewayTestsImage,
		EventServiceImage:                     options.eventServiceImage,
		EventServiceTestsImage:                options.eventServiceTestsImage,
		ApplicationConnectivityValidatorImage: options.applicationConnectivityValidatorImage,
		StrictMode:                            options.strictMode,
	}

	appClient, err := versioned.NewForConfig(cfg)
	if err != nil {
		log.Fatal(err)
	}

	releaseManager := appRelease.NewApplicationReleaseManager(helmClient, appClient.ApplicationconnectorV1alpha1().Applications(), overridesDefaults, options.namespace)

	return releaseManager, nil
}
