package wait

import (
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/waiter"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ForServiceBroker(name, namespace string, svcatCli *clientset.Clientset) error {
	return waiter.WaitAtMost(func() (bool, error) {
		instance, err := svcatCli.ServicecatalogV1beta1().ServiceBrokers(namespace).Get(name, metav1.GetOptions{})
		if err != nil || instance == nil {
			return false, err
		}

		return true, nil
	}, readyTimeout)
}
