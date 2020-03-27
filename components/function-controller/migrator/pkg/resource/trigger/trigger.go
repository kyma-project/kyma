package trigger

import (
	"time"

	"github.com/kyma-project/kyma/components/function-controller/migrator/pkg/resource"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	eventingv1alpha1 "knative.dev/eventing/pkg/apis/eventing/v1alpha1"
)

type Trigger struct {
	resCli      *resource.Resource
	name        string
	namespace   string
	waitTimeout time.Duration
}

func New(dynamicCli dynamic.Interface, name, namespace string, waitTimeout time.Duration, logFn func(format string, args ...interface{})) *Trigger {
	return &Trigger{
		resCli: resource.New(dynamicCli, schema.GroupVersionResource{
			Version:  eventingv1alpha1.SchemeGroupVersion.Version,
			Group:    eventingv1alpha1.SchemeGroupVersion.Group,
			Resource: "triggers",
		}, namespace, logFn),
		name:        name,
		namespace:   namespace,
		waitTimeout: waitTimeout,
	}
}

func (t *Trigger) List(callbacks ...func(...interface{})) ([]*eventingv1alpha1.Trigger, error) {
	ul, err := t.resCli.List(callbacks...)
	if err != nil {
		return nil, err
	}

	var triggers []*eventingv1alpha1.Trigger

	for _, u := range ul.Items {
		var res eventingv1alpha1.Trigger
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.UnstructuredContent(), &res)
		if err != nil {
			return nil, errors.Wrapf(err, "while converting Trigger %s in namespace %s", u.GetName(), u.GetNamespace())
		}

		triggers = append(triggers, &res)
	}

	return triggers, nil
}

func (t *Trigger) Update(res *eventingv1alpha1.Trigger, callbacks ...func(...interface{})) error {
	_, err := t.resCli.Update(res, callbacks...)
	if err != nil {
		return errors.Wrapf(err, "while updating Trigger %s in namespace %s", t.name, t.namespace)
	}

	return nil
}
