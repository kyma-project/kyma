package broker

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/helpers"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/resource"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/shared"

	"github.com/pkg/errors"
)

const DefaultName = "default"

type Broker struct {
	resCli      *resource.Resource
	name        string
	namespace   string
	waitTimeout time.Duration
	log         *logrus.Entry
	verbose     bool
}

func New(c shared.Container) *Broker {
	gvr := schema.GroupVersionResource{
		Group: "eventing.knative.dev", Version: "v1alpha1",
		Resource: "brokers",
	}
	return &Broker{
		resCli:      resource.New(c.DynamicCli, gvr, c.Namespace, c.Log, c.Verbose),
		name:        DefaultName,
		namespace:   c.Namespace,
		waitTimeout: c.WaitTimeout,
		log:         c.Log,
		verbose:     c.Verbose,
	}
}

func (b *Broker) Get() (*unstructured.Unstructured, error) {
	u, err := b.resCli.Get(b.name)
	if err != nil {
		return &unstructured.Unstructured{}, errors.Wrapf(err, "while getting Broker %s in namespace %s", b.name, b.namespace)
	}
	return u, nil
}

func (b *Broker) Delete() error {
	err := b.resCli.Delete(b.name)
	if err != nil {
		return errors.Wrapf(err, "while deleting Broker %s in namespace %s", b.name, b.namespace)
	}

	return nil
}

func (b *Broker) LogResource() error {
	broker, err := b.Get()
	if err != nil {
		return err
	}

	out, err := helpers.PrettyMarshall(broker)
	if err != nil {
		return err
	}

	b.log.Infof("Broker resource: %s", out)
	return nil
}

func (b *Broker) WaitForStatusRunning() error {
	broker, err := b.Get()
	if err != nil {
		return err
	}

	if b.isStateReady(*broker) {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), b.waitTimeout)
	defer cancel()

	condition := b.isBrokerReady(b.name)
	return resource.WaitUntilConditionSatisfied(ctx, b.resCli.ResCli, condition)
}

func (b *Broker) isBrokerReady(name string) func(event watch.Event) (bool, error) {
	return func(event watch.Event) (bool, error) {
		broker, ok := event.Object.(*unstructured.Unstructured)
		if !ok {
			return false, shared.ErrInvalidDataType
		}
		if broker.GetName() != name {
			b.log.Infof("names mismatch, object's name %s, supplied %s", broker.GetName(), name)
			return false, nil
		}

		return b.isStateReady(*broker), nil
	}
}

func (b Broker) isStateReady(broker unstructured.Unstructured) bool {
	conditions, found, err := unstructured.NestedSlice(broker.Object, "status", "conditions")
	if err != nil {
		// status.conditions may have not been added by eventing controller by now
		b.log.Warn("Broker does not have status.conditions")
		return false
	}
	if !found {
		// or it may not even exist, but it should not be the case
		b.log.Warn("Broker not found")
		return false
	}

	ready := false
	for _, cond := range conditions {
		// casting to map[string]string here doesn't work, ok is false
		cond, ok := cond.(map[string]interface{})
		if !ok {
			b.log.Warn("couldn't cast broker's condition to map[string]interface{}")
			ready = false
			break
		}
		if cond["type"].(string) != "Ready" {
			continue
		}

		ready = cond["status"].(string) == "True"

		break
	}

	shared.LogReadiness(ready, b.verbose, b.name, b.log, broker)

	return ready
}
