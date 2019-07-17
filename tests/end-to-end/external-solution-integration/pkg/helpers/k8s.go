package helpers

import (
	"github.com/avast/retry-go"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	coreApi "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
)

const (
	testIDLabel = "testID"
)

var (
	testID string
)

func AddFlags(set *pflag.FlagSet) {
	set.StringVar(&testID, "testID", "e2e", "Unique ID to label all resources")
}

func IsPodReady(pod coreApi.Pod) bool {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == coreApi.PodReady {
			return condition.Status == coreApi.ConditionTrue
		}
	}
	return false
}

func AwaitResourceDeleted(check func() (interface{}, error), opts ...retry.Option) error {
	return retry.Do(func() error {
		_, err := check()

		if err == nil {
			return errors.New("resource still exists")
		}

		if !k8sErrors.IsNotFound(err) {
			return err
		}

		return nil
	}, opts...)
}
