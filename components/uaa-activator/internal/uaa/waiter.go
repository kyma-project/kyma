package uaa

import (
	"context"

	"github.com/kyma-project/kyma/components/uaa-activator/internal/waiter"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	uaaClusterServiceClassName = "xsuaa"
	uaaClusterServicePlanName  = "z54zhz47zdx5loz51z6z58zhvcdz59-b207b177b40ffd4b314b30635590e0ad"
)

// Waiter waits until UAA ClusterServiceClass and ClusterServicePlan are available
type Waiter struct {
	cli client.Client
}

// NewWaiter returns a new instance of Waiter
func NewWaiter(cli client.Client) *Waiter {
	return &Waiter{cli: cli}
}

// WaitForUAAClassAndPlan waits for UAA class and plan
func (w *Waiter) WaitForUAAClassAndPlan(ctx context.Context) error {
	for objName, obj := range map[string]runtime.Object{
		uaaClusterServiceClassName: &v1beta1.ClusterServiceClass{},
		uaaClusterServicePlanName:  &v1beta1.ClusterServicePlan{},
	} {
		if err := w.wait(ctx, objName, obj); err != nil {
			return err
		}
	}

	return nil
}

func (w *Waiter) wait(ctx context.Context, objName string, obj runtime.Object) error {
	key := client.ObjectKey{
		Name: objName,
	}

	objectIsAvailable := func() (err error) {
		return w.cli.Get(ctx, key, obj)
	}

	if err := waiter.WaitForSuccess(ctx, objectIsAvailable); err != nil {
		return errors.Wrapf(err, "while waiting for %s", key.String())
	}

	return nil
}
