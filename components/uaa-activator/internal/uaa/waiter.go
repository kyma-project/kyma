package uaa

import (
	"context"

	"github.com/kyma-project/kyma/components/uaa-activator/internal/repeat"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Waiter waits until UAA ClusterServiceClass and ClusterServicePlan are available
type Waiter struct {
	cli client.Client
	cfg Config
}

// NewWaiter returns a new instance of Waiter
func NewWaiter(cli client.Client, cfg Config) *Waiter {
	return &Waiter{cli: cli, cfg: cfg}
}

// WaitForUAAClassAndPlan waits for UAA class and plan
func (w *Waiter) WaitForUAAClassAndPlan(ctx context.Context) error {
	for objName, obj := range map[string]runtime.Object{
		w.cfg.UAAClusterServiceClassName: &v1beta1.ClusterServiceClass{},
		w.cfg.UAAClusterServicePlanName:  &v1beta1.ClusterServicePlan{},
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

	if err := repeat.UntilSuccess(ctx, objectIsAvailable); err != nil {
		return errors.Wrapf(err, "while waiting for %s", key.String())
	}

	return nil
}
