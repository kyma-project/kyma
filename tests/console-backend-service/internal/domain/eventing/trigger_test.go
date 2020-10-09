// +build acceptance

package eventing

import (
	"fmt"
	"testing"

	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared/auth"

	"github.com/kyma-project/kyma/tests/console-backend-service/internal/configurer"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	tester "github.com/kyma-project/kyma/tests/console-backend-service"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/client"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTriggerEventQueries(t *testing.T) {
	c, err := graphql.New()
	assert.NoError(t, err)

	coreCli, _, err := client.NewClientWithConfig()
	require.NoError(t, err)

	namespace, err := fixNamespace(coreCli)
	require.NoError(t, err)
	namespaceName := namespace.Name()
	defer namespace.Delete()

	_, err = fixService(coreCli, namespaceName)
	require.NoError(t, err)

	subscription := subscribeTriggerEvent(c, createTriggerEventArguments(namespaceName), triggerEventDetailsFields())
	defer subscription.Close()

	err = mutationTrigger(c, "create", createTriggerArguments(namespaceName), triggerDetailsFields())
	assert.NoError(t, err)

	event, err := readTriggerEvent(subscription)
	assert.NoError(t, err)

	expectedEvent := newTriggerEvent("ADD", fixTrigger(namespaceName))
	checkTriggerEvent(t, expectedEvent, event)

	triggers, err := listTriggers(c, listTriggersArguments(namespaceName), triggerDetailsFields())
	assert.NoError(t, err)

	checkTriggerList(t, triggers, TriggerName)

	err = mutationTrigger(c, "delete", deleteTriggerArguments(namespaceName), metadataDetailsFields())
	assert.NoError(t, err)

	triggers, err = listTriggers(c, listTriggersArguments(namespaceName), triggerDetailsFields())
	assert.NoError(t, err)

	checkTriggerList(t, triggers)

	opts := &auth.OperationsInput{
		auth.Create: {fixTriggerRequest("mutation", "createTrigger",
			createTriggerArguments(namespaceName), triggerDetailsFields())},
		auth.List: {fixTriggerRequest("query", "triggers",
			listTriggersArguments(namespaceName), triggerDetailsFields())},
		auth.Delete: {fixTriggerRequest("mutation", "deleteTrigger",
			deleteTriggerArguments(namespaceName), metadataDetailsFields())},
		auth.Watch: {fixTriggerRequest("subscription", "triggerEvent",
			createTriggerEventArguments(namespaceName), triggerEventDetailsFields())},
	}
	auth.New().Run(t, opts)
}

func newTriggerEvent(eventType string, trigger Trigger) TriggerEvent {
	return TriggerEvent{
		Type:    eventType,
		Trigger: trigger,
	}
}

func checkTriggerList(t *testing.T, query TriggerListQueryResponse, expectedNames ...string) {
	assert.Equal(t, len(expectedNames), len(query.Triggers))
	for _, trigger := range query.Triggers {
		assert.Equal(t, true, isInArray(trigger.Name, expectedNames...))
	}
}

func isInArray(data string, array ...string) bool {
	for _, elem := range array {
		if elem == data {
			return true
		}
	}
	return false
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
	fmt.Println("sub next")
	err := sub.Next(&response, tester.DefaultDeletionTimeout)
	fmt.Println(err)

	return response.TriggerEvent, err
}

func fixTriggerRequest(requestType, requestName, arguments, details string) *graphql.Request {
	query := fmt.Sprintf(`
		%s {
			%s (
				%s
			){
				%s
			}
		}
	`, requestType, requestName, arguments, details)
	return graphql.NewRequest(query)
}

func listTriggers(client *graphql.Client, arguments, resourceDetailsQuery string) (TriggerListQueryResponse, error) {
	req := fixTriggerRequest("query", "triggers", arguments, resourceDetailsQuery)
	var res TriggerListQueryResponse
	err := client.Do(req, &res)

	return res, err
}

func mutationTrigger(client *graphql.Client, requestName, arguments, resourceDetailsQuery string) error {
	req := fixTriggerRequest("mutation", fmt.Sprintf("%sTrigger", requestName), arguments, resourceDetailsQuery)
	err := client.Do(req, nil)

	return err
}

func subscribeTriggerEvent(client *graphql.Client, arguments, resourceDetailsQuery string) *graphql.Subscription {
	req := fixTriggerRequest("subscription", "triggerEvent", arguments, resourceDetailsQuery)

	return client.Subscribe(req)
}

func listTriggersArguments(namespace string) string {
	return fmt.Sprintf(`
		namespace: "%s",
		serviceName: "%s"
	`, namespace, SubscriberName)
}

func createTriggerArguments(namespace string) string {
	return fmt.Sprintf(`
		namespace: "%s",
		trigger: {
			name: "%s",
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
	`, namespace, TriggerName, BrokerName, SubscriberAPIVersion, SubscriberKind, SubscriberName, namespace)
}

func createTriggerEventArguments(namespace string) string {
	fmt.Println(fmt.Sprintf(`
		namespace: "%s",
		serviceName: "%s"
	`, namespace, SubscriberName))
	return fmt.Sprintf(`
		namespace: "%s",
		serviceName: "%s"
	`, namespace, SubscriberName)
}

func deleteTriggerArguments(namespace string) string {
	return fmt.Sprintf(`
		namespace: "%s"
		triggerName: "%s"
	`, namespace, TriggerName)
}

func triggerDetailsFields() string {
	return `
		name
		namespace
		spec {
			broker
			filter
			port
			path
			subscriber {
				ref {
					apiVersion
					kind
					name
					namespace
				}
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

func fixTrigger(namespace string) Trigger {
	return Trigger{
		Name:      TriggerName,
		Namespace: namespace,
	}
}

func fixNamespace(dynamicCli *corev1.CoreV1Client) (*configurer.NamespaceConfigurer, error) {
	namespace := configurer.NewNamespace(NamespacePrefix, dynamicCli)
	labels := map[string]string{"knative-eventing-injection": "enabled"}
	err := namespace.Create(labels)

	return namespace, err
}

func fixService(dynamicCli *corev1.CoreV1Client, namespace string) (*configurer.Service, error) {
	service := configurer.NewService(dynamicCli, SubscriberName, namespace)
	spec := v1.ServiceSpec{
		Ports: []v1.ServicePort{
			{
				Protocol:   v1.ProtocolTCP,
				Port:       80,
				TargetPort: intstr.FromInt(8080),
			},
		},
	}
	err := service.Create(spec)

	return service, err
}
