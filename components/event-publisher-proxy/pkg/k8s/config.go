package k8s

import (
	"flag"
	"log"
	"path/filepath"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func ConfigOrDie() *rest.Config {
	c, err := rest.InClusterConfig()
	if err != nil {
		path := getKubeconfigPath()
		c = getConfigOrDie(path)
	}
	return c
}

func getKubeconfigPath() string {
	defaultPath := filepath.Join(homedir.HomeDir(), ".kube", "config")
	path := flag.String("kubeconfig", defaultPath, "Kubeconfig absolute path")
	flag.Parse()
	return *path
}

func getConfigOrDie(kubeconfigPath string) *rest.Config {
	c, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		log.Fatalf("error:[%v]", err)
	}
	return c
}
