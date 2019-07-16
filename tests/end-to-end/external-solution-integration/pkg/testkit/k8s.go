package testkit

import (
	"github.com/avast/retry-go"
	"github.com/pkg/errors"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"

)

type K8sHelper struct {}

func (h K8sHelper) AwaitResourceDeleted(check func() (interface{}, error), opts ...retry.Option) error {
	return retry.Do(func() error {
		_, err := check()

		if err == nil {
			return errors.New("resource still exists")
		}

		if !k8s_errors.IsNotFound(err) {
			return err
		}

		return nil
	}, opts...)
}
