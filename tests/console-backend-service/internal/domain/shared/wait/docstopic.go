package wait

import (
	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	tester "github.com/kyma-project/kyma/tests/console-backend-service"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/waiter"
)

func ForDocsTopicReady(name string, get func(name string) (*v1alpha1.DocsTopic, error)) error {
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
