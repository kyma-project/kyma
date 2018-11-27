package main

import (
	"fmt"
	"os"
	"time"

	"github.com/kyma-project/kyma/components/connection-token-handler/pkg/apis"
	"github.com/kyma-project/kyma/components/connection-token-handler/pkg/controller/tokenrequest"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

func main() {
	logf.SetLogger(logf.ZapLogger(false))
	log := logf.Log.WithName("entrypoint")
	log.Info("Starting ConnectionTokenHandler Controller.")

	options := parseArgs()
	log.Info(fmt.Sprintf("Options: %s", options))

	log.Info("Setting up client for manager")
	cfg, err := config.GetConfig()
	if err != nil {
		log.Error(err, "unable to set up client config")
		os.Exit(1)
	}

	log.Info("Setting up manager")
	syncPeriod := time.Second * time.Duration(options.syncPeriod)
	mgrOpts := manager.Options{
		SyncPeriod: &syncPeriod,
	}
	mgr, err := manager.New(cfg, mgrOpts)
	if err != nil {
		log.Error(err, "unable to set up overall controller manager")
		os.Exit(1)
	}

	log.Info("Registering Components.")

	log.Info("setting up scheme")
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "unable to add APIs to scheme")
		os.Exit(1)
	}

	log.Info("Setting up controllers")
	tokenrequestOpts := tokenrequest.Options{
		TokenTTL:            options.tokenTTL,
		ConnectorServiceURL: options.connectorServiceURL,
	}
	if err := tokenrequest.Add(mgr, tokenrequestOpts); err != nil {
		log.Error(err, "unable to register the tokenrequest controller to the manager")
		os.Exit(1)
	}

	log.Info("Starting the Cmd.")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "unable to run the manager")
		os.Exit(1)
	}
}
