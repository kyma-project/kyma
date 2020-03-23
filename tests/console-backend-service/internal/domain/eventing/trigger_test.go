// +build acceptance

package eventing

import (
	"fmt"
	"testing"

	"github.com/kyma-project/kyma/tests/console-backend-service/internal/resource"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/dynamic"

	tester "github.com/kyma-project/kyma/tests/console-backend-service"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/client"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	Namespace            = "trigger-test"
	TriggerName          = "test-trigger"
	SubscriberName       = "test-subscriber"
	SubscriberAPIVersion = "v1"
	SubscriberKind       = "Service"
	BrokerName           = "default"
)

func TestTriggerEventQueries(t *testing.T) {
	c, err := graphql.New()
	assert.NoError(t, err)

	eventingCli, _, err := client.NewDynamicClientWithConfig()
	require.NoError(t, err)

	//Create namespace
	namespace, err = fixNamespace(eventingCli, t.Logf)
	require.NoError(t, err)

	//Create service
	_, err = fixService(eventingCli, t.Logf)
	require.NoError(t, err)

	//Subscribe events
	subscription := subscribeTriggerEvent(c, createTriggerArguments(), triggerEventDetailsFields())
	defer subscription.Close()

	//Create Trigger
	err = mutationTrigger(c, "create", createTriggerArguments(), triggerDetailsFields())
	assert.NoError(t, err)

	//Check and compare events
	event, err := readTriggerEvent(subscription)
	assert.NoError(t, err)

	expectedEvent := newTriggerEvent("ADD", fixTrigger())
	checkTriggerEvent(t, expectedEvent, event)

	//List triggers
	//err = listTriggers(c, listTriggersArguments(), triggerDetailsFields())
	//assert.NoError(t, err)

	//Delete trigger
	err = mutationTrigger(c, "delete", deleteTriggerArguments(), metadataDetailsFields())

	//Cleanup
	//err = namespace.Delete(Namespace)
	//assert.NoError(err)
}

func newTriggerEvent(eventType string, trigger Trigger) TriggerEvent {
	return TriggerEvent{
		Type:    eventType,
		Trigger: trigger,
	}
}

func checkTriggerEvent(t *testing.T, expected, actual TriggerEvent) {
	assert.Equal(t, expected.Type, actual.Type)
	assert.Equal(t, expected.Trigger.Name, actual.Trigger.Name)
	assert.Equal(t, expected.Trigger.Namespace, actual.Trigger.Namespace)
}

//func checkTriggerList(t *testing.T, triggers []Trigger) {
//
//	assert.Equal(t, trigger.Namespace, TriggerNamespace)
//	assert.Equal(t, trigger.Namespace, TriggerNamespace)
//	assert.Equal(t, trigger.Namespace, TriggerNamespace)
//}

func readTriggerEvent(sub *graphql.Subscription) (TriggerEvent, error) {
	type Response struct {
		TriggerEvent TriggerEvent
	}

	var response Response
	err := sub.Next(&response, tester.DefaultDeletionTimeout)

	return response.TriggerEvent, err
}

func listTriggers(client *graphql.Client, arguments, resourceDetailsQuery string) error {
	query := fmt.Sprintf(`
		query{
			triggers (
				%s
			){
				%s
			}
		}
	`, arguments, resourceDetailsQuery)
	req := graphql.NewRequest(query)
	err := client.Do(req, nil)

	return err
}

func listTriggersArguments() string {
	return fmt.Sprintf(`
		namespace: "%s"
	`, Namespace)
}

func mutationTrigger(client *graphql.Client, requestType, arguments, resourceDetailsQuery string) error {
	query := fmt.Sprintf(`
		mutation {
			%sTrigger (
				%s
			){
				%s
			}
		}
	`, requestType, arguments, resourceDetailsQuery)
	req := graphql.NewRequest(query)
	err := client.Do(req, nil)

	return err
}

func createTriggerArguments() string {
	return fmt.Sprintf(`
		trigger: {
			name: "%s",
			namespace: "%s",
			broker: "%s"
			subscriber: {
				ref: {
					apiVersion: "%s",
					kind: "%s",
					name: "%s",
					namespace: "%s"
				}
			}
		},
	`, TriggerName, Namespace, BrokerName, SubscriberAPIVersion, SubscriberKind, SubscriberName, Namespace)
}

func subscribeTriggerEvent(client *graphql.Client, arguments, resourceDetailsQuery string) *graphql.Subscription {
	query := fmt.Sprintf(`
		subscription {
			triggerEvent (
				%s
			){
				%s
			}
		}
	`, arguments, resourceDetailsQuery)
	req := graphql.NewRequest(query)

	return client.Subscribe(req)
}

func createTriggerArguments() string {
	return fmt.Sprintf(`
		namespace: "%s",
		subscriber: {
			ref: {
				apiVersion: "%s",
				kind: "%s",
				name: "%s",
				namespace: "%s"
			}
		}
	`, Namespace, SubscriberAPIVersion, SubscriberKind, SubscriberName, Namespace)
}

func deleteTriggerArguments() string {
	return fmt.Sprintf(`
		name: "%s",
		namespace: "%s",
	`, TriggerName, Namespace)
}

func triggerDetailsFields() string {
	return `
		name
    	namespace
		broker
    	filterAttributes
		subscriber {
			uri
			ref {
				apiVersion
				kind
				name
				namespace
			}
		}
    	status {
		reason
		status
		}
	`
}

func metadataDetailsFields() string {
	return `
		name
		namespace
	`
}

func triggerEventDetailsFields() string {
	return fmt.Sprintf(`
        type
        trigger {
			%s
        }
    `, triggerDetailsFields())
}

func fixTrigger() Trigger {
	return Trigger{
		Name:      TriggerName,
		Namespace: Namespace,
	}
}

func fixNamespace(dynamicCli dynamic.Interface, logFn func(format string, args ...interface{})) (*resource.Namespace, error) {
	namespace := resource.NewNamespace(dynamicCli, logFn)
	labels := map[string]string{"knative-eventing-injection": "enabled"}
	err := namespace.Create(Namespace, labels)

	return namespace, err
}

func fixService(dynamicCli dynamic.Interface, logFn func(format string, args ...interface{})) (*resource.Service, error) {
	service := resource.NewService(dynamicCli, Namespace, logFn)
	spec := v1.ServiceSpec{
		Ports: []v1.ServicePort{
			{
				Protocol:   v1.ProtocolTCP,
				Port:       80,
				TargetPort: intstr.FromInt(8080),
			},
		},
	}
	err := service.Create(SubscriberName, spec)

	return service, err
}
