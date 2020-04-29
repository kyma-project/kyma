package broker

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	watchtools "k8s.io/client-go/tools/watch"
	eventingv1alpha1 "knative.dev/eventing/pkg/apis/eventing/v1alpha1"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/resource"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/shared"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var (
	ErrInvalidDataType = errors.New("invalid data type")
)

const DefaultBrokerName = "default"

type Broker struct {
	resCli      *resource.Resource
	name        string
	namespace   string
	waitTimeout time.Duration
	log         shared.Logger
}

func New(dynamicCli dynamic.Interface, namespace string, waitTimeout time.Duration, log shared.Logger) *Broker {
	return &Broker{
		resCli: resource.New(dynamicCli, schema.GroupVersionResource{
			Version:  eventingv1alpha1.SchemeGroupVersion.Version,
			Group:    eventingv1alpha1.SchemeGroupVersion.Group,
			Resource: "brokers",
		}, namespace, log),
		name:        DefaultBrokerName,
		namespace:   namespace,
		waitTimeout: waitTimeout,
		log:         log,
	}
}

func (b *Broker) get() (string, error) {
	u, err := b.resCli.Get(b.name)
	if err != nil {
		return "", errors.Wrapf(err, "while getting Broker %s in namespace %s", b.name, b.namespace)
	}

	return u.GetResourceVersion(), nil
}

func (b *Broker) Delete() error {
	err := b.resCli.Delete(b.name, b.waitTimeout)
	if err != nil {
		return errors.Wrapf(err, "while deleting Broker %s in namespace %s", b.name, b.namespace)
	}

	return nil
}

func (b *Broker) WaitForStatusRunning() error {
	resVersion, err := b.get()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), b.waitTimeout)
	defer cancel()
	condition := b.isBrokerReady(b.name)
	_, err = watchtools.Until(ctx, resVersion, b.resCli.ResCli, condition)
	if err != nil {
		return err
	}
	return nil
}

func (b *Broker) isBrokerReady(name string) func(event watch.Event) (bool, error) {
	return func(event watch.Event) (bool, error) {
		if event.Type != watch.Modified {
			return false, nil
		}
		u, ok := event.Object.(*unstructured.Unstructured)
		if !ok {
			return false, ErrInvalidDataType
		}
		if u.GetName() != name {
			return false, nil
		}

		broker := eventingv1alpha1.Broker{}
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &broker)
		if err != nil {
			return false, err
		}

		if broker.Status.IsReady() {
			b.log.Logf("%s is ready:\n%v", name, u)
			return true, nil
		}

		b.log.Logf("%s is not ready:\n%v", name, u)
		return false, nil
	}
}
