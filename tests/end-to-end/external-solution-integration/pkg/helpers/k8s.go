package helpers

import (
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/retry"
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

// IsPodReady checks whether the PodReady condition is true
func IsPodReady(pod v1.Pod) bool {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == v1.PodReady {
			return condition.Status == v1.ConditionTrue
		}
	}
	return false
}

// AwaitResourceDeleted retries until the resources cannot be found any more
func AwaitResourceDeleted(check func() (interface{}, error)) error {
	return retry.Do(func() error {
		_, err := check()

		if err == nil {
			return errors.New("resource still exists")
		}

		if !k8serrors.IsNotFound(err) {
			return err
		}

		return nil
	})
}
