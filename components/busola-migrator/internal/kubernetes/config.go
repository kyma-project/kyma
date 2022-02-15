package kubernetes

import (
	"fmt"
	"os"
	"os/user"
	"path"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func GetKubeConfig() (*rest.Config, error) {
	var restConfig *rest.Config
	var err error

	if kubeEnv := os.Getenv(clientcmd.RecommendedConfigPathEnvVar); len(kubeEnv) > 0 {
		restConfig, err = clientcmd.BuildConfigFromFlags("", kubeEnv)
		if err != nil {
			return nil, err
		}
	} else {
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		if _, ok := os.LookupEnv("HOME"); !ok {
			u, err := user.Current()
			if err != nil {
				return nil, fmt.Errorf("could not get current user: %v", err)
			}
			loadingRules.Precedence = append(loadingRules.Precedence, path.Join(u.HomeDir, clientcmd.RecommendedHomeDir, clientcmd.RecommendedFileName))
		}

		restConfig, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, nil).ClientConfig()
		if err != nil {
			return nil, err
		}
	}

	if err := rest.LoadTLSFiles(restConfig); err != nil {
		return nil, err
	}

	return restConfig, nil
}
