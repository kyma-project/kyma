package main

import (
	"flag"
	"os"
	"path/filepath"
	"strings"

	gatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	oldapi "github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma-project.io/v1alpha2"
	"github.com/kyma-project/kyma/components/api-gateway-migrator/pkg/finder"
	"github.com/kyma-project/kyma/components/api-gateway-migrator/pkg/migrator"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	networkingv1alpha3 "knative.dev/pkg/apis/istio/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = gatewayv1alpha1.AddToScheme(scheme)
	_ = networkingv1alpha3.AddToScheme(scheme)
	_ = rulev1alpha1.AddToScheme(scheme)
	_ = oldapi.AddToScheme(scheme)
}

func main() {
	var omitApisWithLabels string
	flag.StringVar(&omitApisWithLabels, "label-blacklist", "", "Comma-separated list of keys or key=value pairs defining labels of objects that should be omitted. If only a key is provided, then any value will be matched.")
	flag.Parse()
	labels := parseLabels(omitApisWithLabels)

	kubeConfig := initKubeConfig()

	copts := client.Options{
		Scheme: scheme,
	}

	k8sClient, err := client.New(kubeConfig, copts)
	if err != nil {
		log.Fatal(err)
	}

	// TODO: pass config from flags
	defaultConfig := migrator.Config{
		RetriesCount:        3,
		DelayBetweenSteps:   2,
		DelayBetweenRetries: 2,
	}

	clientWrapper := migrator.NewClient(k8sClient, defaultConfig)
	f := finder.New(clientWrapper, labels)

	apisToMigrate, err := f.Find()

	if err != nil {
		log.Fatal(err)
	}

	if len(apisToMigrate) == 0 {
		log.Info("no apis to migrate")
	} else {
		log.Infof("number of apis to migrate: %d", len(apisToMigrate))
	}

	for _, apiToMigrate := range apisToMigrate {
		tmp := apiToMigrate
		m := migrator.New(clientWrapper)
		res, err := m.MigrateOldApi(&tmp)
		if err != nil {
			log.Error(err)
		} else {
			log.Infof("api object: %s/%s successfully migrated to %s\n", tmp.Name, tmp.Namespace, res.NewApiName)
		}
	}

	os.Exit(0)
}

func parseLabels(labelsString string) map[string]*string {
	output := make(map[string]*string)

	for _, labelString := range strings.Split(labelsString, ",") {
		label := strings.TrimSpace(labelString)
		if label == "" {
			continue
		}
		if !strings.Contains(label, "=") {
			output[label] = nil
			continue
		}
		keyValuePair := strings.Split(label, "=")
		output[keyValuePair[0]] = &keyValuePair[1]
	}
	log.Info("labels of objects that will be omitted:")
	if len(output) == 0 {
		log.Info("no labels configured")
	}
	for key, value := range output {
		if value != nil {
			log.Infof(`"%s": "%s"`, key, *value)
		} else {
			log.Infof(`"%s": "<all values>"`, key)
		}
	}
	return output
}

func initKubeConfig() *rest.Config {
	kubeConfigLocation := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigLocation)
	if err != nil {
		log.Warn("unable to build kube config from file. Trying in-cluster configuration")
		kubeConfig, err = rest.InClusterConfig()
		if err != nil {
			log.Fatal("cannot find Service Account in pod to build in-cluster kube config")
		}
	}
	return kubeConfig
}
