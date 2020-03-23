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

func TestTriggerEventQueries(t *testing.T) {
	c, err := graphql.New()
	assert.NoError(t, err)

	eventingCli, _, err := client.NewDynamicClientWithConfig()
	require.NoError(t, err)

	namespace, err := fixNamespace(eventingCli, t.Logf)
	require.NoError(t, err)

	_, err = fixService(eventingCli, t.Logf)
	require.NoError(t, err)

	subscription := subscribeTriggerEvent(c, createTriggerEventArguments(), triggerEventDetailsFields())
	defer subscription.Close()

	err = mutationTrigger(c, "create", createTriggerArguments(), triggerDetailsFields())
	assert.NoError(t, err)

	event, err := readTriggerEvent(subscription)
	assert.NoError(t, err)

	expectedEvent := newTriggerEvent("ADD", fixTrigger())
	checkTriggerEvent(t, expectedEvent, event)

	//List triggers
	//err = listTriggers(c, listTriggersArguments(), triggerDetailsFields())
	//assert.NoError(t, err)

	err = mutationTrigger(c, "delete", deleteTriggerArguments(), metadataDetailsFields())
	assert.NoError(t, err)

	err = namespace.Delete(Namespace)
	assert.NoError(t, err)
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

func createTriggerEventArguments() string {
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
		trigger: {
			name: "%s",
			namespace: "%s",
		}
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
