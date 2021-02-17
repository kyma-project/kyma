package trigger

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"
	watchtools "k8s.io/client-go/tools/watch"
	duckv1 "knative.dev/pkg/apis/duck/v1"

	"github.com/sirupsen/logrus"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/broker"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/helpers"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/resource"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/shared"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	eventingv1alpha1 "knative.dev/eventing/pkg/apis/eventing/v1alpha1"
)

type Trigger struct {
	resCli      *resource.Resource
	name        string
	namespace   string
	waitTimeout time.Duration
	log         *logrus.Entry
	verbose     bool
}

func New(name string, c shared.Container) *Trigger {
	return &Trigger{
		resCli:      resource.New(c.DynamicCli, eventingv1alpha1.SchemeGroupVersion.WithResource("triggers"), c.Namespace, c.Log, c.Verbose),
		name:        name,
		namespace:   c.Namespace,
		waitTimeout: c.WaitTimeout,
		log:         c.Log,
		verbose:     c.Verbose,
	}
}

func (t *Trigger) Create(serviceName string) error {
	br := &eventingv1alpha1.Trigger{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Trigger",
			APIVersion: eventingv1alpha1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      t.name,
			Namespace: t.namespace,
		},
		Spec: eventingv1alpha1.TriggerSpec{
			Broker: broker.DefaultName,
			Subscriber: duckv1.Destination{
				Ref: &corev1.ObjectReference{
					Kind:       "Service",
					Namespace:  t.namespace,
					Name:       serviceName,
					APIVersion: corev1.SchemeGroupVersion.Version,
				},
			},
		},
	}

	_, err := t.resCli.Create(br)
	if err != nil {
		return errors.Wrapf(err, "while creating Trigger %s in namespace %s", t.name, t.namespace)
	}

	return err
}

func (t *Trigger) Delete() error {
	err := t.resCli.Delete(t.name, t.waitTimeout)
	if err != nil {
		return errors.Wrapf(err, "while deleting Trigger %s in namespace %s", t.name, t.namespace)
	}

	return nil
}

func (t *Trigger) Get() (*eventingv1alpha1.Trigger, error) {
	u, err := t.resCli.Get(t.name)
	if err != nil {
		return &eventingv1alpha1.Trigger{}, errors.Wrapf(err, "while getting Trigger %s in namespace %s", t.name, t.namespace)
	}

	tr, err := convertFromUnstructuredToTrigger(*u)
	if err != nil {
		return &eventingv1alpha1.Trigger{}, err
	}

	return &tr, nil
}

func (t *Trigger) LogResource() error {
	trigger, err := t.Get()
	if err != nil {
		return err
	}

	out, err := helpers.PrettyMarshall(trigger)
	if err != nil {
		return err
	}

	t.log.Infof("%s", out)
	return nil
}

func (t *Trigger) WaitForStatusRunning() error {
	tr, err := t.Get()
	if err != nil {
		return err
	}

	if t.isStateReady(*tr) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), t.waitTimeout)
	defer cancel()
	condition := t.isTriggerReady(t.name)
	_, err = watchtools.Until(ctx, tr.GetResourceVersion(), t.resCli.ResCli, condition)
	if err != nil {
		return err
	}
	return nil
}

func (t *Trigger) isTriggerReady(name string) func(event watch.Event) (bool, error) {
	return func(event watch.Event) (bool, error) {
		if event.Type != watch.Modified {
			return false, nil
		}
		u, ok := event.Object.(*unstructured.Unstructured)
		if !ok {
			return false, shared.ErrInvalidDataType
		}
		if u.GetName() != name {
			return false, nil
		}

		trigger, err := convertFromUnstructuredToTrigger(*u)
		if err != nil {
			return false, err
		}

		return t.isStateReady(trigger), nil
	}
}

func convertFromUnstructuredToTrigger(u unstructured.Unstructured) (eventingv1alpha1.Trigger, error) {
	trigger := eventingv1alpha1.Trigger{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &trigger)
	return trigger, err
}

func (t Trigger) isStateReady(trigger eventingv1alpha1.Trigger) bool {
	ready := trigger.Status.IsReady()

	shared.LogReadiness(ready, t.verbose, t.name, t.log, trigger)

	return ready
}
