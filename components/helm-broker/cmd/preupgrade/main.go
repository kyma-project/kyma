package main

import (
	"flag"
	"strings"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

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

	deployments, err := clientset.AppsV1().Deployments(cfg.Namespace).List(v1.ListOptions{
		LabelSelector: "app=helm-broker",
	})
	fatalOnError(err)

	reposURLs := ""
	for _, deploy := range deployments.Items {
		for _, container := range deploy.Spec.Template.Spec.Containers {
			for _, env := range container.Env {
				if env.Name == "URLs" {
					reposURLs = env.Value
				}
			}
		}
	}
	newURLs := strings.Join(strings.Split(reposURLs, ";"), "\n")

	_, err = clientset.CoreV1().ConfigMaps(cfg.Namespace).Create(fixBundlesRepos(newURLs))
	fatalOnError(err)
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

func fixBundlesRepos(urls string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		TypeMeta: v1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      "migrated-bundles-repos",
			Namespace: "",
		},
		Data: map[string]string{
			"URLs": urls,
		},
	}
}
