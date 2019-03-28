package wait

import (
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	tester "github.com/kyma-project/kyma/tests/console-backend-service"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/waiter"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ForServiceInstanceReady(instanceName, namespace string, svcatCli *clientset.Clientset) error {
	return waiter.WaitAtMost(func() (bool, error) {
		instance, err := svcatCli.ServicecatalogV1beta1().ServiceInstances(namespace).Get(instanceName, metav1.GetOptions{})
		if err != nil || instance == nil {
			return false, err
		}

		conditions := instance.Status.Conditions
		for _, cond := range conditions {
			if cond.Type == v1beta1.ServiceInstanceConditionReady {
				return cond.Status == v1beta1.ConditionTrue, nil
			}
		}

		return false, nil
	}, tester.DefaultReadyTimeout)
}

func ForServiceInstanceDeletion(instanceName, namespace string, svcatCli *clientset.Clientset) error {
	return waiter.WaitAtMost(func() (bool, error) {
		_, err := svcatCli.ServicecatalogV1beta1().ServiceInstances(namespace).Get(instanceName, metav1.GetOptions{})

		if errors.IsNotFound(err) {
			return true, nil
		}
		if err != nil {
			return false, err
		}

		return false, nil
	}, tester.DefaultReadyTimeout)
}
