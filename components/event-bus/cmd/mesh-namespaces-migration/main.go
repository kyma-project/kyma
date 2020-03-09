package main

import (
	"context"
	"flag"
	"log"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	// allow client authentication against GKE clusters
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	servicecatalogclientset "github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset"

	kneventingclientset "knative.dev/eventing/pkg/client/clientset/versioned"
	"knative.dev/pkg/injection/sharedmain"

	kymaeventingclientset "github.com/kyma-project/kyma/components/event-bus/client/generated/clientset/internalclientset"
)

const defaultTimeoutDuration = 5 * time.Minute

// Configuration flags
var kubeConfig string

func init() {
	flag.StringVar(&kubeConfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
}

func main() {
	flag.Parse()

	k8sClient, kymaClient, knativeClient, servicecatalogClient := initClientSets()

	userNamespaces, err := listUserNamespaces(k8sClient)
	if err != nil {
		log.Fatalf("Error listing user namespaces: %s", err)
	}

	// initialize managers

	serviceInstanceManager, err := newServiceInstanceManager(servicecatalogClient, kymaClient, userNamespaces)
	if err != nil {
		log.Fatalf("Error initializing serviceInstanceManager: %s", err)
	}

	subscriptionMigrator, err := newSubscriptionMigrator(kymaClient, knativeClient, userNamespaces)
	if err != nil {
		log.Fatalf("Error initializing subscriptionMigrator: %s", err)
	}

	// run migration

	if err := serviceInstanceManager.recreateAll(); err != nil {
		log.Fatalf("Error re-creating ServiceInstances: %s", err)
	}

	if err := subscriptionMigrator.migrateAll(); err != nil {
		log.Fatalf("Error migrating Kyma Subscriptions: %s", err)
	}
}

// initClientSets initializes all required Kubernetes ClientSets.
func initClientSets() (*kubernetes.Clientset,
	*kymaeventingclientset.Clientset,
	*kneventingclientset.Clientset,
	*servicecatalogclientset.Clientset) {

	cfg := getRESTConfig()

	return kubernetes.NewForConfigOrDie(cfg),
		kymaeventingclientset.NewForConfigOrDie(cfg),
		kneventingclientset.NewForConfigOrDie(cfg),
		servicecatalogclientset.NewForConfigOrDie(cfg)
}

// getRESTConfig returns a rest.Config to be used for Kubernetes client creation.
func getRESTConfig() *rest.Config {
	cfg, err := sharedmain.GetConfig("", kubeConfig)
	if err != nil {
		log.Fatal("Error building kubeconfig", err)
	}
	return cfg
}

// excludedNamespaces returns the namespaces to be excluded from the migration.
func excludedNamespaces() map[string]interface{} {
	return map[string]interface{}{
		"istio-system":     "",
		"knative-eventing": "",
		"knative-serving":  "",
		"kube-node-lease":  "",
		"kube-public":      "",
		"kube-system":      "",
		"kyma-installer":   "",
		"kyma-integration": "",
		"kyma-system":      "",
		"natss":            "",
	}
}

// listUserNamespaces returns a list of user namespaces to be migrated.
func listUserNamespaces(k8sClient kubernetes.Interface) ([]string, error) {
	var userNamespaces []string

	namespaces, err := k8sClient.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, ns := range namespaces.Items {
		if _, isExcluded := excludedNamespaces()[ns.Name]; !isExcluded {
			userNamespaces = append(userNamespaces, ns.Name)
		}
	}

	return userNamespaces, nil
}

// newTimeoutChannel returns a channel that receives a value after a default timeout.
func newTimeoutChannel() <-chan struct{} {
	ctx, _ := context.WithTimeout(context.TODO(), defaultTimeoutDuration)
	return ctx.Done()
}
