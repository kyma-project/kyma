package wait

import (
	"github.com/kyma-project/kyma/tests/console-backend-service"
	"github.com/kyma-project/kyma/components/helm-broker/pkg/apis/addons/v1alpha1"
)

func ForAddonsConfigurationStatus(name string, expStatus v1alpha1.AddonsConfigurationPhase, get func(name string) (*v1alpha1.AddonsConfiguration, error)) error {
	return waiter.WaitAtMost(func() (bool, error) {
		res, err := get(name)
		if err != nil {
			return false, err
		}

		if res.Status.Phase == v1alpha1.DocsTopicReady {
			return true, nil
		}

		return false, nil
	}, 2*tester.DefaultReadyTimeout)
}
