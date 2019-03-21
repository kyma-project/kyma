// +build acceptance

package servicecatalog

import (
	"fmt"
	"testing"

	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared/auth"

	tester "github.com/kyma-project/kyma/tests/console-backend-service"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared/fixture"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared/wait"

	"github.com/kyma-project/kyma/tests/console-backend-service/internal/client"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	serviceBindingStatusTypeUnknown = "UNKNOWN"
)

type ServiceBindingEvent struct {
	Type           string
	ServiceBinding shared.ServiceBinding
}

type CreateServiceBindingOutput struct {
	Name                string
	ServiceInstanceName string
	Namespace           string
}

type DeleteServiceBindingOutput struct {
	Name      string
	Namespace string
}

type bindingQueryResponse struct {
	ServiceBinding shared.ServiceBinding
}

type bindingCreateMutationResponse struct {
	CreateServiceBinding CreateServiceBindingOutput
}

type bindingDeleteMutationResponse struct {
	DeleteServiceBinding DeleteServiceBindingOutput
}

func TestServiceBindingMutationsAndQueries(t *testing.T) {
	c, err := graphql.New()
	require.NoError(t, err)

	svcatCli, _, err := client.NewServiceCatalogClientWithConfig()
	require.NoError(t, err)

	instanceName := "binding-test-instance"
	instance := fixture.ServiceInstanceFromClusterServiceClass(instanceName, TestNamespace)

	bindingName := "test-binding"
	binding := fixture.ServiceBinding(bindingName, instanceName, TestNamespace)
	createBindingOutput := createBindingOutput(bindingName, instanceName)
	deleteBindingOutput := deleteBindingOutput(bindingName)

	t.Log("Subscribe bindings")
	subscription := subscribeBinding(c, bindingEventDetailsFields(), binding.Namespace)
	defer subscription.Close()

	t.Log("Create Instance")
	_, err = createInstance(c, "name", instance, true)
	require.NoError(t, err)

	err = wait.ForServiceInstanceReady(instance.Name, instance.Namespace, svcatCli)
	require.NoError(t, err)

	t.Log("Create Binding")
	createRes, err := createBinding(c, createBindingOutput)

	assert.NoError(t, err)
	checkCreateBindingOutput(t, createBindingOutput, createRes.CreateServiceBinding)

	t.Log("Check subscription event")
	expectedEvent := bindingEvent("ADD", binding)
	event, err := readServiceBindingEvent(subscription)
	assert.NoError(t, err)
	checkBindingEvent(t, expectedEvent, event)

	err = wait.ForServiceBindingReady(binding.Name, binding.Namespace, svcatCli)
	require.NoError(t, err)

	t.Log("Query Single Resource")
	res, err := queryBinding(c, binding)

	assert.NoError(t, err)
	checkBinding(t, binding, res.ServiceBinding)

	t.Log("Query Binding of Instance")
	instanceRes, err := querySingleInstance(c, `
		name
		serviceBindings {
			items {
				name
				serviceInstanceName
				namespace
				secret {
					name
					namespace
					data
				}
				status {
					type
				}
			}
			stats {	
				ready
				failed
				pending
				unknown
			}
		}
	`, instance)

	assert.NoError(t, err)
	assertBindingExistsAndEqual(t, binding, instanceRes.ServiceInstance.ServiceBindings.Items)
	stats := instanceRes.ServiceInstance.ServiceBindings.Stats
	assert.Equal(t, 1, stats.Ready+stats.Pending+stats.Failed+stats.Unknown, instanceRes.ServiceInstance.ServiceBindings)

	t.Log("Delete Binding")
	deleteRes, err := deleteBinding(c, deleteBindingOutput)

	assert.NoError(t, err)
	checkDeleteBindingOutput(t, deleteBindingOutput, deleteRes.DeleteServiceBinding)

	t.Log("Wait for binding deletion")
	err = wait.ForServiceBindingDeletion(binding.Name, binding.Namespace, svcatCli)
	require.NoError(t, err)

	t.Log("Delete Instance")
	_, err = deleteInstance(c, "name", instance)
	assert.NoError(t, err)

	t.Log("Wait for instance deletion")
	err = wait.ForServiceInstanceDeletion(instance.Name, instance.Namespace, svcatCli)
	assert.NoError(t, err)

	t.Log("Checking authorization directives...")
	ops := &auth.OperationsInput{
		auth.Get:    {fixQueryBindingRequest(binding)},
		auth.Create: {fixCreateBingindRequest(createBindingOutput)},
		auth.Delete: {fixDeleteBindingRequest(deleteBindingOutput)},
	}
	AuthSuite.Run(t, ops)
}

func subscribeBinding(c *graphql.Client, resourceDetailsQuery string, namespace string) *graphql.Subscription {
	query := fmt.Sprintf(`
			subscription ($namespace: String!) {
				serviceBindingEvent(namespace: $namespace) {
					%s
				}
			}
		`, resourceDetailsQuery)
	req := graphql.NewRequest(query)
	req.SetVar("namespace", namespace)

	return c.Subscribe(req)
}

func fixCreateBingindRequest(expectedResource CreateServiceBindingOutput) *graphql.Request {
	query := `
			mutation ($bindingName: String!, $namespace: String!, $instanceName: String!) {
				createServiceBinding(serviceBindingName: $bindingName, serviceInstanceName: $instanceName, namespace: $namespace) {
					name
					serviceInstanceName
					namespace
				}
			}
	`
	req := graphql.NewRequest(query)
	req.SetVar("bindingName", expectedResource.Name)
	req.SetVar("namespace", expectedResource.Namespace)
	req.SetVar("instanceName", expectedResource.ServiceInstanceName)

	return req
}

func createBinding(c *graphql.Client, expectedResource CreateServiceBindingOutput) (bindingCreateMutationResponse, error) {
	req := fixCreateBingindRequest(expectedResource)

	var res bindingCreateMutationResponse
	err := c.Do(req, &res)

	return res, err
}

func fixQueryBindingRequest(expectedResource shared.ServiceBinding) *graphql.Request {
	query := `
		query ($name: String!, $namespace: String!) {
			serviceBinding(name: $name, namespace: $namespace) {
				name
				serviceInstanceName
				namespace
				secret {
					name
					namespace
					data
				}
				status {
					type
				}
			}
		}
	`
	req := graphql.NewRequest(query)
	req.SetVar("name", expectedResource.Name)
	req.SetVar("namespace", expectedResource.Namespace)

	return req
}

func queryBinding(c *graphql.Client, expectedResource shared.ServiceBinding) (bindingQueryResponse, error) {
	req := fixQueryBindingRequest(expectedResource)

	var res bindingQueryResponse
	err := c.Do(req, &res)

	return res, err
}

func fixDeleteBindingRequest(expectedResource DeleteServiceBindingOutput) *graphql.Request {
	query := `
			mutation ($bindingName: String!, $namespace: String!) {
				deleteServiceBinding(serviceBindingName: $bindingName, namespace: $namespace) {
					name
					namespace
				}
			}
	`
	req := graphql.NewRequest(query)
	req.SetVar("bindingName", expectedResource.Name)
	req.SetVar("namespace", expectedResource.Namespace)

	return req
}

func deleteBinding(c *graphql.Client, expectedResource DeleteServiceBindingOutput) (bindingDeleteMutationResponse, error) {
	req := fixDeleteBindingRequest(expectedResource)

	var res bindingDeleteMutationResponse
	err := c.Do(req, &res)

	return res, err
}

func checkBinding(t *testing.T, expected, actual shared.ServiceBinding) {
	assert.Equal(t, expected.Name, actual.Name)
	assert.Equal(t, expected.Namespace, actual.Namespace)
	assert.Equal(t, expected.ServiceInstanceName, actual.ServiceInstanceName)
	assert.Equal(t, expected.Secret.Name, actual.Secret.Name)
	assert.Equal(t, expected.Secret.Namespace, actual.Secret.Namespace)

	assert.NotEmpty(t, actual.Status.Type)
	assert.NotEqual(t, serviceBindingStatusTypeUnknown, actual.Status.Type)
}

func assertBindingExistsAndEqual(t *testing.T, expectedElement shared.ServiceBinding, arr []shared.ServiceBinding) {
	assert.Condition(t, func() (success bool) {
		for _, v := range arr {
			if v.Name == expectedElement.Name {
				checkBinding(t, expectedElement, v)
				return true
			}
		}

		return false
	}, "Resource does not exist")
}

func checkCreateBindingOutput(t *testing.T, expected, actual CreateServiceBindingOutput) {
	assert.Equal(t, expected.Name, actual.Name)
	assert.Equal(t, expected.Namespace, actual.Namespace)
	assert.Equal(t, expected.ServiceInstanceName, actual.ServiceInstanceName)
}

func checkDeleteBindingOutput(t *testing.T, expected, actual DeleteServiceBindingOutput) {
	assert.Equal(t, expected.Name, actual.Name)
	assert.Equal(t, expected.Namespace, actual.Namespace)
}

func bindingEvent(eventType string, binding shared.ServiceBinding) ServiceBindingEvent {
	return ServiceBindingEvent{
		Type:           eventType,
		ServiceBinding: binding,
	}
}

func createBindingOutput(bindingName, instanceName string) CreateServiceBindingOutput {
	return CreateServiceBindingOutput{
		Name:                bindingName,
		Namespace:           TestNamespace,
		ServiceInstanceName: instanceName,
	}
}

func deleteBindingOutput(bindingName string) DeleteServiceBindingOutput {
	return DeleteServiceBindingOutput{
		Name:      bindingName,
		Namespace: TestNamespace,
	}
}

func bindingEventDetailsFields() string {
	return `
        type
        serviceBinding {
			name
        }
    `
}

func readServiceBindingEvent(sub *graphql.Subscription) (ServiceBindingEvent, error) {
	type Response struct {
		ServiceBindingEvent ServiceBindingEvent
	}
	var bindingEvent Response
	err := sub.Next(&bindingEvent, tester.DefaultSubscriptionTimeout)

	return bindingEvent.ServiceBindingEvent, err
}

func checkBindingEvent(t *testing.T, expected, actual ServiceBindingEvent) {
	assert.Equal(t, expected.Type, actual.Type)
	assert.Equal(t, expected.ServiceBinding.Name, actual.ServiceBinding.Name)
}
