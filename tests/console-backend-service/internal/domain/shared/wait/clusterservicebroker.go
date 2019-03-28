package wait

import (
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/waiter"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ForClusterServiceBrokerReady(name string, svcatCli *clientset.Clientset) error {
	return waiter.WaitAtMost(func() (bool, error) {
		broker, err := svcatCli.ServicecatalogV1beta1().ClusterServiceBrokers().Get(name, metav1.GetOptions{})
		if err != nil || broker == nil {
			return false, err
		}

		for _, v := range broker.Status.Conditions {
			if v.Type == "Ready" && v.Status == "True" {
				return true, nil
			}
		}

		return false, nil
	}, readyTimeout)
}
