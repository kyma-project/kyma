package trigger

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/helpers"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/resource"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/shared"

	"github.com/pkg/errors"
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
	gvr := schema.GroupVersionResource{
		Group: "eventing.knative.dev", Version: "v1alpha1",
		Resource: "triggers",
	}
	return &Trigger{
		resCli:      resource.New(c.DynamicCli, gvr, c.Namespace, c.Log, c.Verbose),
		name:        name,
		namespace:   c.Namespace,
		waitTimeout: c.WaitTimeout,
		log:         c.Log,
		verbose:     c.Verbose,
	}
}

func (t *Trigger) Create(serviceName string) error {
	tr := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "eventing.knative.dev/v1alpha1",
			"kind":       "Trigger",
			"metadata": map[string]interface{}{
				"name":      t.name,
				"namespace": t.namespace,
			},
			"spec": map[string]interface{}{
				"broker": "default",
				"subscriber": map[string]interface{}{
					"ref": map[string]interface{}{
						"kind":       "Service",
						"name":       serviceName,
						"namespace":  t.namespace,
						"apiVersion": corev1.SchemeGroupVersion.Version,
					},
				},
			},
		},
	}

	_, err := t.resCli.Create(tr)
	if err != nil {
		return errors.Wrapf(err, "while creating Trigger %s in namespace %s", t.name, t.namespace)
	}

	return err
}

func (t *Trigger) Delete() error {
	err := t.resCli.Delete(t.name)
	if err != nil {
		return errors.Wrapf(err, "while deleting Trigger %s in namespace %s", t.name, t.namespace)
	}

	return nil
}

func (t *Trigger) Get() (*unstructured.Unstructured, error) {
	tr, err := t.resCli.Get(t.name)
	if err != nil {
		return &unstructured.Unstructured{}, errors.Wrapf(err, "while getting Trigger %s in namespace %s", t.name, t.namespace)
	}
	return tr, nil
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
	return resource.WaitUntilConditionSatisfied(ctx, t.resCli.ResCli, condition)
}

func (t *Trigger) isTriggerReady(name string) func(event watch.Event) (bool, error) {
	return func(event watch.Event) (bool, error) {
		if event.Type != watch.Modified {
			return false, nil
		}
		trigger, ok := event.Object.(*unstructured.Unstructured)
		if !ok {
			return false, shared.ErrInvalidDataType
		}
		if trigger.GetName() != name {
			return false, nil
		}

		return t.isStateReady(*trigger), nil
	}
}

func (t Trigger) isStateReady(trigger unstructured.Unstructured) bool {
	conditions, found, err := unstructured.NestedSlice(trigger.Object, "status", "conditions")
	if err != nil {
		// status.conditions may have not been added by eventing controller by now
		t.log.Warn("Trigger does not have status.conditions")
		return false
	}
	if !found {
		// or it may not even exist, but it should not be the case
		t.log.Warn("Trigger not found")
		return false
	}

	ready := false
	for _, cond := range conditions {
		// casting to map[string]string here doesn't work, ok is false
		cond, ok := cond.(map[string]interface{})
		if !ok {
			t.log.Warn("couldn't cast trigger's condition to map[string]interface{}")
			ready = false
			break
		}
		if cond["type"].(string) != "Ready" {
			continue
		}

		ready = cond["status"].(string) == "True"
		break
	}

	shared.LogReadiness(ready, t.verbose, t.name, t.log, trigger)

	return ready
}
