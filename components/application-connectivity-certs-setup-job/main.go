package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

func main() {

	log.Info("Starting Application Connectivity Certificates setup job")
	options := parseArgs()
	log.Infof("Options: %s", options)

	k8sConfig, err := initKubeConfig()
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

	err = migrateSecrets(secretRepo, *options)
	if err != nil {
		log.Fatalf("Failed to migrate secrets, %s", err.Error())
	}

	certSetupHandler := NewCertificateSetupHandler(options, secretRepo)

	err = certSetupHandler.SetupApplicationConnectorCertificate()
	if err != nil {
		log.Fatalf("Failed to setup certificates, %s", err.Error())
	}

	log.Info("Certificates set up successfully")
}

// Returns intiliazed config, allows local usage (outside cluster) based on provided kubeconfig or in-cluter
func initKubeConfig() (*restclient.Config, error) {

	kubeConfig, err := restclient.InClusterConfig()
	if err == nil {
		return kubeConfig, nil
	}

	log.Warnf("Failed to read in-cluster go client configuration")
	var kubeconfig string

	if kubeconfigPath := os.Getenv("KUBECONFIG"); kubeconfigPath != "" {
		kubeconfig = kubeconfigPath
	} else if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	} else {
		return nil, errors.New("cannot find kubeconfig file")
	}

	return clientcmd.BuildConfigFromFlags("", kubeconfig)
}

func migrateSecrets(secretRepo SecretRepository, options options) error {
	err := migrateSecret(secretRepo, options.caCertificateSecretToMigrate, options.caCertificateSecret, options.caCertificateSecretKeysToMigrate)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to migrate secret %s : %v", options.caCertificateSecretToMigrate, err))
	}

	err = migrateSecret(secretRepo, options.connectorCertificateSecretToMigrate, options.connectorCertificateSecret, options.connectorCertificateSecretKeysToMigrate)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to migrate secret %s : %v", options.connectorCertificateSecretToMigrate, err))
	}

	return nil
}

func migrateSecret(secretRepo SecretRepository, sourceSecret, targetSecret types.NamespacedName, keysToInclude string) error {
	unmarshallKeysList := func(keys string) (keysArray []string, err error) {
		err = json.Unmarshal([]byte(keys), &keysArray)

		return keysArray, err
	}

	keys, err := unmarshallKeysList(keysToInclude)
	if err != nil {
		log.Errorf("Failed to read secret keys to be migrated")
		return err
	}

	migrator := getMigrator(secretRepo, keys)

	return migrator.Do(sourceSecret, targetSecret)
}

func getMigrator(secretRepo SecretRepository, keysToInclude []string) migrator {
	getIncludeSourceKeyFunc := func() IncludeKeyFunc {
		if len(keysToInclude) == 0 {
			return func(string) bool {
				return true
			}
		}

		return func(key string) bool {
			for _, k := range keysToInclude {
				if k == key {
					return true
				}
			}

			return false
		}
	}

	return NewMigrator(secretRepo, getIncludeSourceKeyFunc())
}
