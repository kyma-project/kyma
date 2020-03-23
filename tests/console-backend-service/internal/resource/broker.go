package resource

import (
	"github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type broker struct {
	resCli *Resource
}

func NewBroker(dynamicCli dynamic.Interface, logFn func(format string, args ...interface{})) *broker {
	return &broker{
		resCli: New(dynamicCli, schema.GroupVersionResource{
			Version:  "v1alpha1",
			Group:    "eventing.knative.dev",
			Resource: "brokers",
		}, "", logFn),
	}
}

func (self *broker) Create(name string) error {
	broker := v1alpha1.Broker{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}

	err := self.resCli.Create(broker)
	if err != nil {
		return errors.Wrapf(err, "while creating broker %s", name)
	}

	return err
}

func (self *broker) Get(name string) (*v1alpha1.Broker, error) {
	u, err := self.resCli.Get(name)
	if err != nil {
		return nil, err
	}

	var res v1alpha1.Broker
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &res)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting broker %s", name)
	}

	return &res, nil
}

func (self *broker) Delete(name string) error {
	err := self.resCli.Delete(name)
	if err != nil {
		return errors.Wrapf(err, "while deleting broker %s", name)
	}

	return nil
}
