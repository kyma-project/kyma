package main

import (
	"flag"
	"strings"

	"github.com/kyma-project/kyma/components/helm-broker/platform/logger"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	cfgMapName = "migrated-bundles-repos"
)

// pre-upgrade kyma release 0.7 -> 0.8
func main() {
	verbose := flag.Bool("verbose", false, "specify if log verbosely loading configuration")
	flag.Parse()
	cfg, err := Load(*verbose)
	fatalOnError(err)

	// creates the in-cluster k8sConfig
	k8sConfig, err := newRestClientConfig(cfg.KubeconfigPath)
	fatalOnError(err)

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(k8sConfig)
	fatalOnError(err)

	log := logger.New(&cfg.Logger)

	log.Info("Searching for additional repositories")
	deployments, err := clientset.AppsV1().Deployments(cfg.Namespace).List(v1.ListOptions{
		LabelSelector: "app=helm-broker",
	})
	fatalOnError(err)
	if len(deployments.Items) > 1 {
		log.Warn("Found more then one helm-broker deployments")
	}

	reposURLs := ""
	for _, deploy := range deployments.Items {
		for _, container := range deploy.Spec.Template.Spec.Containers {
			for _, env := range container.Env {
				if env.Name == "APP_REPOSITORY_URLS" {
					reposURLs = env.Value
				}
			}
		}
	}

	newURLs := ""
	for _, repo := range strings.Split(reposURLs, ";") {
		if repo == "https://github.com/kyma-project/bundles/releases/download/0.3.0/" || repo == "https://github.com/kyma-project/bundles/releases/download/0.3.0/index.yaml" {
			continue
		}
		newURLs += repo + "\n"
	}

	if len(newURLs) < 1 {
		log.Info("Not found any repositories")
		return
	}

	log.Infof("Found repositories: %s", newURLs)
	migratedRepos := migratePreviousBundlesRepos(newURLs)
	configMaps := clientset.CoreV1().ConfigMaps(cfg.Namespace)

	_, err = configMaps.Create(migratedRepos)
	switch {
	case err == nil:
	case apiErrors.IsAlreadyExists(err):
		cfgMap, err := configMaps.Get(cfgMapName, v1.GetOptions{})
		fatalOnError(err)

		cfgMap.Data["URLs"] = newURLs
		_, err = configMaps.Update(cfgMap)
		fatalOnError(err)

		log.Infof("Updated migrated bundles config map with following URLs: %s", newURLs)
	default:
		logrus.Fatal(errors.Wrap(err, "while creating ConfigMap with migrated bundles repos"))
	}
}

func fatalOnError(err error) {
	if err != nil {
		logrus.Fatal(err.Error())
	}
}

func newRestClientConfig(kubeConfigPath string) (*rest.Config, error) {
	if kubeConfigPath != "" {
		return clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	}

	return rest.InClusterConfig()
}

func migratePreviousBundlesRepos(urls string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		TypeMeta: v1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: cfgMapName,
			Labels: map[string]string{
				"helm-broker-repo": "true",
			},
		},
		Data: map[string]string{
			"URLs": urls,
		},
	}
}
