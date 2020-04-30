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
	"k8s.io/client-go/dynamic"
	eventingv1alpha1 "knative.dev/eventing/pkg/apis/eventing/v1alpha1"
)

type Trigger struct {
	resCli      *resource.Resource
	name        string
	namespace   string
	waitTimeout time.Duration
	log         shared.Logger
}

type ResourceVersion string

func New(dynamicCli dynamic.Interface, name, namespace string, waitTimeout time.Duration, log shared.Logger, verbose bool) *Trigger {
	return &Trigger{
		resCli: resource.New(dynamicCli, schema.GroupVersionResource{
			Version:  eventingv1alpha1.SchemeGroupVersion.Version,
			Group:    eventingv1alpha1.SchemeGroupVersion.Group,
			Resource: "triggers",
		}, namespace, log, verbose),
		name:        name,
		namespace:   namespace,
		waitTimeout: waitTimeout,
		log:         log,
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
			Broker: broker.DefaultBrokerName,
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

	if tr.Status.IsReady() {
		t.log.Logf("Trigger %s in namespace %s is ready", tr.Name, tr.Namespace)
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

		trigger := eventingv1alpha1.Trigger{}
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &trigger)
		if err != nil {
			return false, err
		}

		if trigger.Status.IsReady() {
			t.log.Logf("Trigger %s is ready", name)
			return true, nil
		}

		t.log.Logf("Trigger %s is not ready", name)
		return false, nil
	}
}
