package main

import (
	"github.com/kyma-project/kyma/components/remote-environment-controller/pkg/api"
	"github.com/kyma-project/kyma/components/remote-environment-controller/pkg/controller"
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

	log.Info("Starting Remote Environment Controller.")

	options := parseArgs()
	log.Infof("Options: %s", options)

	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatal(err)
	}

	mgr, err := manager.New(cfg, manager.Options{})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Registering Components:")

	log.Printf("Setting up scheme.")
	if err := api.AddToScheme(mgr.GetScheme()); err != nil {
		log.Fatal(err)
	}

	log.Printf("Setting up controller.")

	overridesData := controller.OverridesData{
		DomainName:             options.domainName,
		ProxyServiceImage:      options.proxyServiceImage,
		EventServiceImage:      options.eventServiceImage,
		EventServiceTestsImage: options.eventServiceTestsImage,
	}

	err = controller.InitRemoteEnvironmentController(mgr, overridesData, options.namespace, options.appName, options.tillerUrl)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Starting the Cmd.")
	log.Info(mgr.Start(signals.SetupSignalHandler()))
}
