package testsuite

import (
	"fmt"
	"github.com/avast/retry-go"
	"github.com/kyma-project/kyma/tests/end-to-end/external-solution-integration/pkg/step"
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1alpha12 "knative.dev/eventing/pkg/apis/eventing/v1alpha1"
	"knative.dev/eventing/pkg/client/clientset/versioned/typed/eventing/v1alpha1"
	"knative.dev/pkg/apis"
	apisv1alpha1 "knative.dev/pkg/apis/v1alpha1"
)

type CreateKnativeTrigger struct {
	triggers v1alpha1.TriggerInterface
	name     string
	endpoint string
	broker   string
}

func (c CreateKnativeTrigger) Name() string {
	return fmt.Sprintf("Create Knative Trigger for %s broker", c.broker)
}

func (c CreateKnativeTrigger) Run() error {
	url, err := apis.ParseURL(c.endpoint)
	if err != nil {
		return err
	}

	trigger := &v1alpha12.Trigger{
		TypeMeta: v1.TypeMeta{},
		ObjectMeta: v1.ObjectMeta{
			Name: c.name,
		},
		Spec: v1alpha12.TriggerSpec{
			Broker: c.broker,
			Subscriber: &apisv1alpha1.Destination{
				URI: url,
			},
		},
	}
	_, error := c.triggers.Create(trigger)
	if error != nil {
		return error
	}

	return retry.Do(c.isKnativeTriggerReady)
}

func (c CreateKnativeTrigger) isKnativeTriggerReady() error {
	trigger, err := c.triggers.Get(c.name, v1.GetOptions{})
	if err != nil {
		return err
	}
	if !trigger.Status.IsReady() {
		return errors.Errorf("knative trigger with name: %s is not ready. Status of Knative Tigger:\n %v", c.name, trigger.Status)
	}
	return nil
}

func (c CreateKnativeTrigger) Cleanup() error {
	return c.triggers.Delete(c.name, &v1.DeleteOptions{})
}

var _ step.Step = &CreateKnativeTrigger{}

//NewCreateKnativeTrigger returns new CreateKnativeTrigger
func NewCreateKnativeTrigger(triggerName, brokerName, endpoint string, trigger v1alpha1.TriggerInterface) *CreateKnativeTrigger {
	return &CreateKnativeTrigger{
		triggers: trigger,
		name:     triggerName,
		endpoint: endpoint,
		broker:   brokerName,
	}
}
