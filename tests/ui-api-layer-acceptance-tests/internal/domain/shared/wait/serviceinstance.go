package wait

import (
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/internal/waiter"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

const (
	readyTimeout = time.Second * 45
)

func ForServiceInstanceReady(instanceName, environment string, svcatCli *clientset.Clientset) error {
	return waiter.WaitAtMost(func() (bool, error) {
		instance, err := svcatCli.ServicecatalogV1beta1().ServiceInstances(environment).Get(instanceName, metav1.GetOptions{})
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
	}, readyTimeout)
}

func ForServiceInstanceDeletion(instanceName, environment string, svcatCli *clientset.Clientset) error {
	return waiter.WaitAtMost(func() (bool, error) {
		_, err := svcatCli.ServicecatalogV1beta1().ServiceInstances(environment).Get(instanceName, metav1.GetOptions{})

		if errors.IsNotFound(err) {
			return true, nil
		}
		if err != nil {
			return false, err
		}

		return false, nil
	}, readyTimeout)
}
