package framework

import (
	"flag"
	"fmt"
	"strings"

	arkclient "github.com/heptio/ark/pkg/generated/clientset/versioned"
	kubelesscli "github.com/kubeless/kubeless/pkg/client/clientset/versioned"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	sbuClient "github.com/kyma-project/kyma/components/binding-usage-controller/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/tests/backup-restore-e2e/consts"
	"github.com/kyma-project/kyma/tests/backup-restore-e2e/utils"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"

	// needed for kubeconfigs for gke
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	"k8s.io/client-go/tools/clientcmd"
)

// Global - instance of Framework
var Global *Framework

// Framework contains clients for several components
type Framework struct {
	KubeClient     kubernetes.Interface
	REClient       versioned.Interface
	SBClient       clientset.Interface
	ArkClient      arkclient.Interface
	KubelessClient kubelesscli.Interface
	SbuClient      sbuClient.Interface
	Namespace      string
	Deprovision    bool
	Brokers        []string
}

func setup() error {
	logrus.Info("Setting up framework...")
	kubeconfig := flag.String("kubeconfig", "", "kube config path, e.g. $HOME/.kube/config")
	ns := flag.String("namespace", "default", "e2e test namespace")
	deprovision := flag.Bool("deprovision", false, "If deprovision of Instances/Bindings should be done")
	brokers := flag.String("brokers", fmt.Sprintf("%s,%s", consts.BrokerReb, consts.BrokerHelm), "Provide comma separated list of brokers which instances/bindings should be tested")
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)

	if err != nil {
		return err
	}

	defaultCli, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	reCli, err := versioned.NewForConfig(config)
	if err != nil {
		return err
	}

	sbCli, err := clientset.NewForConfig(config)
	if err != nil {
		return err
	}

	arkCli, err := arkclient.NewForConfig(config)
	if err != nil {
		return err
	}

	kubelessCli, err := kubelesscli.NewForConfig(config)
	if err != nil {
		return err
	}

	sbuCli, err := sbuClient.NewForConfig(config)
	if err != nil {
		return err
	}

	Global = &Framework{
		KubeClient:     defaultCli,
		REClient:       reCli,
		SBClient:       sbCli,
		ArkClient:      arkCli,
		KubelessClient: kubelessCli,
		SbuClient:      sbuCli,
		Namespace:      *ns,
		Deprovision:    *deprovision,
		Brokers:        strings.Split(*brokers, ","),
	}

	return nil
}

func teardown() error {

	if Global.Deprovision {
		logrus.Info("Deprovisioning of Service Instances & Bindings")
		utils.DeleteAllServiceBindings(Global.Namespace, Global.SBClient)
		utils.DeleteAllServiceInstances(Global.Namespace, Global.SBClient)

		logrus.Info("Waiting until all Service Bindings are deleted")
		errb := utils.WaitForAllBindingsDeleted(Global.SBClient, Global.Namespace)
		logrus.Info("Waiting until all Service Instances are deleted")
		erri := utils.WaitForAllInstancesDeleted(Global.SBClient, Global.Namespace)

		if errb != nil || erri != nil {
			return fmt.Errorf("Error during deprovisioning: %v, %v", errb, erri)
		}
	}

	return nil
}

// IsBrokerSelected returns true if provided broker was passed in --brokers argument
func IsBrokerSelected(broker string) bool {
	for _, ae := range Global.Brokers {
		if ae == broker {
			return true
		}
	}
	return false
}
