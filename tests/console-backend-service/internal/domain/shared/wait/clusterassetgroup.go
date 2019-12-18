package wait

import (
	tester "github.com/kyma-project/kyma/tests/console-backend-service"
	"github.com/kyma-project/kyma/tests/console-backend-service/pkg/waiter"
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
)

func ForClusterAssetGroupReady(name string, get func(name string) (*v1beta1.ClusterAssetGroup, error)) error {
	return waiter.WaitAtMost(func() (bool, error) {
		res, err := get(name)
		if err != nil {
			return false, err
		}

		if res.Status.Phase == v1beta1.AssetGroupReady {
			return true, nil
		}

		return false, nil
	}, 2*tester.DefaultReadyTimeout)
}

func ForClusterAssetGroupDeletion(name string, get func(name string) (*v1beta1.ClusterAssetGroup, error)) error {
	return waiter.WaitAtMost(func() (bool, error) {
		_, err := get(name)
		if errors.IsNotFound(err) {
			return true, nil
		}
		if err != nil {
			return false, err
		}
		return true, nil
	}, 2*tester.DefaultReadyTimeout)
}
