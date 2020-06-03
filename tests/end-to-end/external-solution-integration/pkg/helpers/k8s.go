package helpers

import (
	"fmt"

	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/retry"
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

// IsDeploymentReady checks whether the DeploymentReady condition is true
func IsDeploymentReady(deployment appsv1.Deployment) bool {
	for _, condition := range deployment.Status.Conditions {
		if condition.Type == appsv1.DeploymentAvailable {
			return condition.Status == v1.ConditionTrue
		}
	}
	return false
}

// InClusterEndpoint build in-cluster address to communicate with services
func InClusterEndpoint(name, namespace string, port int) string {
	return fmt.Sprintf("http://%s.%s.svc.cluster.local:%v", name, namespace, port)
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
