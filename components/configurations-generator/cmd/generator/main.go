package main

import (
	"encoding/base64"
	"flag"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/configurations-generator/pkg/kube_config"
	log "github.com/sirupsen/logrus"
)

func main() {

	port := flag.Int("port", 8000, "Application port")
	clusterNameArg := flag.String("kube-config-custer-name", "", "Name of the Kubernetes cluster")
	urlArg := flag.String("kube-config-url", "", "URL of the Kubernetes Apiserver")
	caArg := flag.String("kube-config-ca", "", "Certificate Authority of the Kubernetes cluster")
	caFileArg := flag.String("kube-config-ca-file", "", "File with Certificate Authority of the Kubernetes cluster")
	namespaceArg := flag.String("kube-config-ns", "", "Default namespace of the Kubernetes context")
	flag.Parse()

	log.Infof("Starting configurations generator on port: %d...", *port)

	kubeConfig := KubeConfigFromArgs(clusterNameArg, urlArg, caArg, caFileArg, namespaceArg)
	kubeConfigEndpoints := kube_config.NewEndpoints(kubeConfig)

	router := mux.NewRouter()
	router.Methods("GET").Path("/kube-config").HandlerFunc(kubeConfigEndpoints.GetKubeConfig)

	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(*port), router))
}

func KubeConfigFromArgs(clusterNameArg, urlArg, caArg, caFileArg, namespaceArg *string) *kube_config.KubeConfig {

	var clusterName, url, ca, namespace string

	if clusterNameArg == nil || *clusterNameArg == "" {
		log.Fatal("Name of the Kubernetes cluster is required.")
	} else {
		clusterName = *clusterNameArg
	}

	if urlArg == nil || *urlArg == "" {
		log.Fatal("URL of the Kubernetes Apiserver is required.")
	} else {
		url = *urlArg
	}

	if caArg == nil || *caArg == "" {
		ca = readCaFromFile(caFileArg)
	} else {
		ca = *caArg
	}

	if namespaceArg != nil {
		namespace = *namespaceArg
	}

	return kube_config.NewKubeConfig(clusterName, url, ca, namespace)
}

func readCaFromFile(caFileArg *string) string {

	if caFileArg == nil || *caFileArg == "" {
		log.Fatal("Certificate Authority of the Kubernetes cluster is required.")
	}

	caBytes, caErr := ioutil.ReadFile(*caFileArg)
	if caErr != nil {
		log.Fatalf("Error while reading Certificate Authority of the Kubernetes cluster. Root cause: %v", caErr)
	}

	return base64.StdEncoding.EncodeToString(caBytes)
}
