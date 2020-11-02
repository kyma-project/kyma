package broker

import (
	"context"
	"time"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/helpers"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	watchtools "k8s.io/client-go/tools/watch"
	eventingv1alpha1 "knative.dev/eventing/pkg/apis/eventing/v1alpha1"

	"github.com/sirupsen/logrus"

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
	return &Broker{
		resCli:      resource.New(c.DynamicCli, eventingv1alpha1.SchemeGroupVersion.WithResource("brokers"), c.Namespace, c.Log, c.Verbose),
		name:        DefaultName,
		namespace:   c.Namespace,
		waitTimeout: c.WaitTimeout,
		log:         c.Log,
		verbose:     c.Verbose,
	}
}

func (b *Broker) Get() (*eventingv1alpha1.Broker, error) {
	u, err := b.resCli.Get(b.name)
	if err != nil {
		return &eventingv1alpha1.Broker{}, errors.Wrapf(err, "while getting Broker %s in namespace %s", b.name, b.namespace)
	}

	broker, err := convertFromUnstructuredToBroker(*u)
	if err != nil {
		return &eventingv1alpha1.Broker{}, err
	}

	return &broker, nil
}

func (b *Broker) Delete() error {
	err := b.resCli.Delete(b.name, b.waitTimeout)
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
	_, err = watchtools.Until(ctx, broker.GetResourceVersion(), b.resCli.ResCli, condition)
	if err != nil {
		return err
	}
	return nil
}

func (b *Broker) isBrokerReady(name string) func(event watch.Event) (bool, error) {
	return func(event watch.Event) (bool, error) {
		u, ok := event.Object.(*unstructured.Unstructured)
		if !ok {
			return false, shared.ErrInvalidDataType
		}
		if u.GetName() != name {
			b.log.Infof("names mismatch, object's name %s, supplied %s", u.GetName(), name)
			return false, nil
		}

		broker, err := convertFromUnstructuredToBroker(*u)
		if err != nil {
			return false, err
		}

		return b.isStateReady(broker), nil
	}
}

func convertFromUnstructuredToBroker(u unstructured.Unstructured) (eventingv1alpha1.Broker, error) {
	broker := eventingv1alpha1.Broker{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &broker)
	return broker, err
}

func (b Broker) isStateReady(broker eventingv1alpha1.Broker) bool {
	ready := broker.Status.IsReady()

	shared.LogReadiness(ready, b.verbose, b.name, b.log, broker)

	return ready
}
