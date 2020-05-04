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
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/broker"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/resource"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/shared"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	eventingv1alpha1 "knative.dev/eventing/pkg/apis/eventing/v1alpha1"
)

type Trigger struct {
	resCli      *resource.Resource
	name        string
	namespace   string
	waitTimeout time.Duration
	log         shared.Logger
	verbose     bool
}

type ResourceVersion string

func New(name string, c shared.Container) *Trigger {
	return &Trigger{
		resCli: resource.New(c.DynamicCli, schema.GroupVersionResource{
			Version:  eventingv1alpha1.SchemeGroupVersion.Version,
			Group:    eventingv1alpha1.SchemeGroupVersion.Group,
			Resource: "triggers",
		}, c.Namespace, c.Log, c.Verbose),
		name:        name,
		namespace:   c.Namespace,
		waitTimeout: c.WaitTimeout,
		log:         c.Log,
		verbose:     c.Verbose,
	}
}

func (t *Trigger) Create(serviceName string) (ResourceVersion, error) {
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
					APIVersion: servingv1.SchemeGroupVersion.String(),
				},
			},
		},
	}

	resourceVersion, err := t.resCli.Create(br)
	if err != nil {
		return ResourceVersion(resourceVersion), errors.Wrapf(err, "while creating Trigger %s in namespace %s", t.name, t.namespace)
	}

	return ResourceVersion(resourceVersion), err
}

func (t *Trigger) Delete() error {
	err := t.resCli.Delete(t.name, t.waitTimeout)
	if err != nil {
		return errors.Wrapf(err, "while deleting Trigger %s in namespace %s", t.name, t.namespace)
	}

	return nil
}

func (t *Trigger) get() (*eventingv1alpha1.Trigger, error) {
	u, err := t.resCli.Get(t.name)
	if err != nil {
		return &eventingv1alpha1.Trigger{}, errors.Wrapf(err, "while getting Trigger %s in namespace %s", t.name, t.namespace)
	}

	trigger := &eventingv1alpha1.Trigger{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, trigger)
	if err != nil {
		return &eventingv1alpha1.Trigger{}, err
	}

	return trigger, nil
}

func (t *Trigger) WaitForStatusRunning(initialResourceVersion ResourceVersion) error {
	tr, err := t.get()
	if err != nil {
		return err
	}

	if t.isStateReady(*tr) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), t.waitTimeout)
	defer cancel()
	condition := t.isTriggerReady(t.name)
	_, err = watchtools.Until(ctx, string(initialResourceVersion), t.resCli.ResCli, condition)
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

	if ready {
		t.log.Logf("Trigger %s is ready", t.name)
	} else {
		t.log.Logf("Trigger %s is not ready", t.name)
	}

	if t.verbose {
		t.log.Logf("%+v", trigger)
	}

	return ready
}
