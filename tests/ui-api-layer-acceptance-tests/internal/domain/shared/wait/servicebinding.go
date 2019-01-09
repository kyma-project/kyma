package wait

import (
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/internal/waiter"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ForServiceBindingReady(name, environment string, svcatCli *clientset.Clientset) error {
	return waiter.WaitAtMost(func() (bool, error) {
		instance, err := svcatCli.ServicecatalogV1beta1().ServiceBindings(environment).Get(name, metav1.GetOptions{})
		if err != nil || instance == nil {
			return false, err
		}

		arr := instance.Status.Conditions
		for _, v := range arr {
			if v.Type == "Ready" {
				return v.Status == "True", nil
			}
		}

		return false, nil
	}, readyTimeout)
}

func ForServiceBindingDeletion(name, environment string, svcatCli *clientset.Clientset) error {
	return waiter.WaitAtMost(func() (bool, error) {
		_, err := svcatCli.ServicecatalogV1beta1().ServiceBindings(environment).Get(name, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			return true, nil
		}
		if err != nil {
			return false, err
		}
		return true, nil
	}, readyTimeout)
}
