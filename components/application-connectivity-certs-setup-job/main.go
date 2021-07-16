package main

import (
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

func main() {

	log.Info("Starting Application Connectivity Certificates setup job")
	options := parseArgs()
	log.Infof("Options: %s", options)

	k8sConfig, err := restclient.InClusterConfig()
	if err != nil {
		log.Fatalf("Failed to get in cluster config, %s", err.Error())
	}

	coreClientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		log.Fatalf("Failed to create core client set, %s", err.Error())
	}

	secretRepo := NewSecretRepository(func(namespace string) Manager {
		return coreClientset.CoreV1().Secrets(namespace)
	})

	migrator := NewMigrator(secretRepo)
	migrator.Do(options.caCertificateSecretToMigrate, options.caCertificateSecret)

	certSetupHandler := NewCertificateSetupHandler(options, secretRepo)

	err = certSetupHandler.SetupApplicationConnectorCertificate()
	if err != nil {
		log.Fatalf("Failed to setup certificates, %s", err.Error())
	}

	log.Info("Certificates set up successfully")
}
