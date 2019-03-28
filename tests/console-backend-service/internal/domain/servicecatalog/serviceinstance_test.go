// +build acceptance

package servicecatalog

import (
	"fmt"
	"testing"

	tester "github.com/kyma-project/kyma/tests/console-backend-service"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/client"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared/fixture"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared/wait"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ServiceInstanceEvent struct {
	Type            string
	ServiceInstance shared.ServiceInstance
}

type instancesQueryResponse struct {
	ServiceInstances []shared.ServiceInstance
}

type instanceQueryResponse struct {
	ServiceInstance shared.ServiceInstance
}

type instanceCreateMutationResponse struct {
	CreateServiceInstance shared.ServiceInstance
}

type instanceDeleteMutationResponse struct {
	DeleteServiceInstance shared.ServiceInstance
}

func TestServiceInstanceMutationsAndQueries(t *testing.T) {
	c, err := graphql.New()
	require.NoError(t, err)

	svcatCli, _, err := client.NewServiceCatalogClientWithConfig()
	require.NoError(t, err)

	expectedResourceFromClusterServiceClass := fixture.ServiceInstance("cluster-test-instance", TestNamespace)
	expectedResourceFromServiceClass := instanceFromServiceClass("test-instance")
	resourceDetailsQuery := instanceDetailsFields()

	t.Log(fmt.Sprintf("Subscribe instance created by %s", tester.ClusterServiceBroker))
	subscription := subscribeInstance(c, instanceEventDetailsFields(), expectedResourceFromClusterServiceClass.Namespace)
	defer subscription.Close()

	t.Log(fmt.Sprintf("Create instance from %s", tester.ClusterServiceBroker))
	createRes, err := createInstance(c, resourceDetailsQuery, expectedResourceFromClusterServiceClass, true)

	require.NoError(t, err)
	checkInstanceFromClusterServiceClass(t, expectedResourceFromClusterServiceClass, createRes.CreateServiceInstance)

	t.Log(fmt.Sprintf("Check subscription event of instance created by %s", tester.ClusterServiceBroker))
	expectedEvent := instanceEvent("ADD", expectedResourceFromClusterServiceClass)
	event, err := readInstanceEvent(subscription)
	assert.NoError(t, err)
	checkInstanceEvent(t, expectedEvent, event)

	t.Log(("Wait for instance Ready created by %s"), tester.ClusterServiceBroker)
	err = wait.ForServiceInstanceReady(expectedResourceFromClusterServiceClass.Name, expectedResourceFromClusterServiceClass.Namespace, svcatCli)
	assert.NoError(t, err)

	t.Log(fmt.Sprintf("Create instance from %s", tester.ServiceBroker))
	createRes, err = createInstance(c, resourceDetailsQuery, expectedResourceFromServiceClass, false)

	require.NoError(t, err)
	checkInstanceFromServiceClass(t, expectedResourceFromServiceClass, createRes.CreateServiceInstance)

	t.Log(fmt.Sprintf("Wait for instance Ready created by %s", tester.ServiceBroker))
	err = wait.ForServiceInstanceReady(expectedResourceFromServiceClass.Name, expectedResourceFromServiceClass.Namespace, svcatCli)
	assert.NoError(t, err)

	t.Log(fmt.Sprintf("Query Single Resource - instance created by %s", tester.ClusterServiceBroker))
	res, err := querySingleInstance(c, resourceDetailsQuery, expectedResourceFromClusterServiceClass)

	assert.NoError(t, err)
	checkInstanceFromClusterServiceClass(t, expectedResourceFromClusterServiceClass, res.ServiceInstance)

	t.Log(fmt.Sprintf("Query Single Resource - instance created by %s", tester.ServiceBroker))
	res, err = querySingleInstance(c, resourceDetailsQuery, expectedResourceFromServiceClass)

	assert.NoError(t, err)
	checkInstanceFromServiceClass(t, expectedResourceFromServiceClass, res.ServiceInstance)

	t.Log("Query Multiple Resources")
	multipleRes, err := queryMultipleInstances(c, resourceDetailsQuery, TestNamespace)

	assert.NoError(t, err)
	assertInstanceFromClusterServiceClassExistsAndEqual(t, expectedResourceFromClusterServiceClass, multipleRes.ServiceInstances)
	assertInstanceFromServiceClassExistsAndEqual(t, expectedResourceFromServiceClass, multipleRes.ServiceInstances)

	// We must again wait for RUNNING status of created instances, because sometimes Kubernetess change status from RUNNING to PROVISIONING at the first queries - Query Single Resource
	t.Log(fmt.Sprintf("Wait for instance Ready created by %s", tester.ClusterServiceBroker))
	err = wait.ForServiceInstanceReady(expectedResourceFromClusterServiceClass.Name, expectedResourceFromClusterServiceClass.Namespace, svcatCli)
	assert.NoError(t, err)

	t.Log(fmt.Sprintf("Wait for instance Ready created by %s", tester.ServiceBroker))
	err = wait.ForServiceInstanceReady(expectedResourceFromServiceClass.Name, expectedResourceFromServiceClass.Namespace, svcatCli)
	assert.NoError(t, err)

	t.Log("Query Multiple Resources With Status")
	multipleResWithStatus, err := queryMultipleInstancesWithStatus(c, resourceDetailsQuery, TestNamespace)

	assert.NoError(t, err)
	assertInstanceFromClusterServiceClassExistsAndEqual(t, expectedResourceFromClusterServiceClass, multipleResWithStatus.ServiceInstances)
	assertInstanceFromServiceClassExistsAndEqual(t, expectedResourceFromServiceClass, multipleRes.ServiceInstances)

	t.Log(fmt.Sprintf("Delete instance created by %s", tester.ClusterServiceBroker))
	deleteRes, err := deleteInstance(c, resourceDetailsQuery, expectedResourceFromClusterServiceClass)

	assert.NoError(t, err)
	checkInstanceFromClusterServiceClass(t, expectedResourceFromClusterServiceClass, deleteRes.DeleteServiceInstance)

	t.Log(fmt.Sprintf("Wait for deletion of instance created by %s", tester.ClusterServiceBroker))
	err = wait.ForServiceInstanceDeletion(expectedResourceFromClusterServiceClass.Name, expectedResourceFromClusterServiceClass.Namespace, svcatCli)
	assert.NoError(t, err)

	t.Log(fmt.Sprintf("Delete instance created by %s", tester.ServiceBroker))
	deleteRes, err = deleteInstance(c, resourceDetailsQuery, expectedResourceFromServiceClass)

	assert.NoError(t, err)
	checkInstanceFromServiceClass(t, expectedResourceFromServiceClass, deleteRes.DeleteServiceInstance)

	t.Log(fmt.Sprintf("Wait for deletion of instance created by %s", tester.ServiceBroker))
	err = wait.ForServiceInstanceDeletion(expectedResourceFromServiceClass.Name, expectedResourceFromServiceClass.Namespace, svcatCli)
	assert.NoError(t, err)
}

func createInstance(c *graphql.Client, resourceDetailsQuery string, expectedResource shared.ServiceInstance, clusterWide bool) (instanceCreateMutationResponse, error) {
	query := fmt.Sprintf(`
			mutation ($name: String!, $namespace: String!, $externalPlanName: String!, $externalServiceClassName: String!, $labels: [String!]!, $parameterSchema: JSON) {
				createServiceInstance(params: {
    				name: $name,
    				namespace: $namespace,
					classRef: {
						externalName: $externalServiceClassName,
						clusterWide: %v,
					},
					planRef: {
						externalName: $externalPlanName,
						clusterWide: %v,
					},
    				labels: $labels,
					parameterSchema: $parameterSchema
				}) {
					%s
				}
			}
		`, clusterWide, clusterWide, resourceDetailsQuery)
	req := graphql.NewRequest(query)
	req.SetVar("name", expectedResource.Name)
	req.SetVar("namespace", expectedResource.Namespace)
	if clusterWide {
		req.SetVar("externalPlanName", expectedResource.ClusterServicePlan.ExternalName)
		req.SetVar("externalServiceClassName", expectedResource.ClusterServiceClass.ExternalName)
	} else {
		req.SetVar("externalPlanName", expectedResource.ServicePlan.ExternalName)
		req.SetVar("externalServiceClassName", expectedResource.ServiceClass.ExternalName)
	}
	req.SetVar("labels", expectedResource.Labels)
	req.SetVar("parameterSchema", expectedResource.PlanSpec)

	var res instanceCreateMutationResponse
	err := c.Do(req, &res)

	return res, err
}

func subscribeInstance(c *graphql.Client, resourceDetailsQuery string, namespace string) *graphql.Subscription {
	query := fmt.Sprintf(`
			subscription ($namespace: String!) {
				serviceInstanceEvent(namespace: $namespace) {
					%s
				}
			}
		`, resourceDetailsQuery)
	req := graphql.NewRequest(query)
	req.SetVar("namespace", namespace)

	return c.Subscribe(req)
}

func querySingleInstance(c *graphql.Client, resourceDetailsQuery string, expectedResource shared.ServiceInstance) (instanceQueryResponse, error) {
	req := singleResourceQueryRequest(resourceDetailsQuery, expectedResource)

	var res instanceQueryResponse
	err := c.Do(req, &res)

	return res, err
}

func queryMultipleInstances(c *graphql.Client, resourceDetailsQuery, namespace string) (instancesQueryResponse, error) {
	query := fmt.Sprintf(`
			query ($namespace: String!) {
				serviceInstances(namespace: $namespace) {
					%s
				}
			}	
		`, resourceDetailsQuery)
	req := graphql.NewRequest(query)
	req.SetVar("namespace", namespace)

	var res instancesQueryResponse
	err := c.Do(req, &res)

	return res, err
}

func queryMultipleInstancesWithStatus(c *graphql.Client, resourceDetailsQuery, namespace string) (instancesQueryResponse, error) {
	query := fmt.Sprintf(`
			query ($namespace: String!, $status: InstanceStatusType) {
				serviceInstances(namespace: $namespace, status: $status) {
					%s
				}
			}	
		`, resourceDetailsQuery)
	req := graphql.NewRequest(query)
	req.SetVar("namespace", namespace)
	req.SetVar("status", shared.ServiceInstanceStatusTypeRunning)

	var res instancesQueryResponse
	err := c.Do(req, &res)

	return res, err
}

func deleteInstance(c *graphql.Client, resourceDetailsQuery string, expectedResource shared.ServiceInstance) (instanceDeleteMutationResponse, error) {
	query := fmt.Sprintf(`
			mutation ($name: String!, $namespace: String!) {
				deleteServiceInstance(name: $name, namespace: $namespace) {
					%s
				}
			}
		`, resourceDetailsQuery)
	req := graphql.NewRequest(query)
	req.SetVar("name", expectedResource.Name)
	req.SetVar("namespace", expectedResource.Namespace)

	var res instanceDeleteMutationResponse
	err := c.Do(req, &res)

	return res, err
}

func singleResourceQueryRequest(resourceDetailsQuery string, expectedResource shared.ServiceInstance) *graphql.Request {
	query := fmt.Sprintf(`
			query ($name: String!, $namespace: String!) {
				serviceInstance(name: $name, namespace: $namespace) {
					%s
				}
			}
		`, resourceDetailsQuery)
	req := graphql.NewRequest(query)
	req.SetVar("name", expectedResource.Name)
	req.SetVar("namespace", expectedResource.Namespace)

	return req
}

func instanceDetailsFields() string {
	return `
		name
		namespace
		planSpec
		bindable
		creationTimestamp
		labels
		classReference {
			name
			displayName
			clusterWide	
		}
		planReference {
			name
			displayName
			clusterWide
		}
		status {
			type
			reason
			message
		}
		clusterServicePlan {
			name
			displayName
			externalName
			description
			relatedClusterServiceClassName
			instanceCreateParameterSchema
		}
		clusterServiceClass {
			name
			externalName
			displayName
			creationTimestamp
			description
			longDescription
			imageUrl
			documentationUrl
			supportUrl
			providerDisplayName
			tags
			activated
		}
		servicePlan {
			name
			displayName
			externalName
			description
			relatedServiceClassName
			instanceCreateParameterSchema
			bindingCreateParameterSchema
		 }
		 serviceClass {
			name
			namespace
			externalName
			displayName
			creationTimestamp
			description
			longDescription
			imageUrl
			documentationUrl
			supportUrl
			providerDisplayName
			tags
			activated
		}
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
			}
			stats {
				ready
				failed
				pending
				unknown
			}
        }
		serviceBindingUsages {
			name
			namespace
			usedBy {
				kind
				name
			}
		}
	`
}

func instanceEventDetailsFields() string {
	return fmt.Sprintf(`
        type
        serviceInstance {
			%s
        }
    `, instanceDetailsFields())
}

func checkInstanceFromClusterServiceClass(t *testing.T, expected, actual shared.ServiceInstance) {
	// Name
	assert.Equal(t, expected.Name, actual.Name)

	// Namespace
	assert.Equal(t, expected.Namespace, actual.Namespace)

	// ClusterServicePlan.Name
	assert.Equal(t, expected.ClusterServicePlan.Name, actual.ClusterServicePlan.Name)

	// ClusterServiceClass.Name
	assert.Equal(t, expected.ClusterServiceClass.Name, actual.ClusterServiceClass.Name)
	assert.Equal(t, expected.Labels, actual.Labels)
	assert.Equal(t, expected.Bindable, actual.Bindable)
}

func checkInstanceFromServiceClass(t *testing.T, expected, actual shared.ServiceInstance) {
	// Name
	assert.Equal(t, expected.Name, actual.Name)

	// Namespace
	assert.Equal(t, expected.Namespace, actual.Namespace)

	// ServicePlan.Name
	assert.Equal(t, expected.ServicePlan.Name, actual.ServicePlan.Name)

	// ServiceClass.Name
	assert.Equal(t, expected.ServiceClass.Name, actual.ServiceClass.Name)

	// ServiceClass.Namespace
	assert.Equal(t, expected.ServiceClass.Namespace, actual.ServiceClass.Namespace)
}

func assertInstanceFromClusterServiceClassExistsAndEqual(t *testing.T, expectedElement shared.ServiceInstance, arr []shared.ServiceInstance) {
	assert.Condition(t, func() (success bool) {
		for _, v := range arr {
			if v.Name == expectedElement.Name {
				checkInstanceFromClusterServiceClass(t, expectedElement, v)
				return true
			}
		}

		return false
	}, "Resource does not exist")
}

func assertInstanceFromServiceClassExistsAndEqual(t *testing.T, expectedElement shared.ServiceInstance, arr []shared.ServiceInstance) {
	assert.Condition(t, func() (success bool) {
		for _, v := range arr {
			if v.Name == expectedElement.Name {
				checkInstanceFromServiceClass(t, expectedElement, v)
				return true
			}
		}

		return false
	}, "Resource does not exist")
}

func instanceFromServiceClass(name string) shared.ServiceInstance {
	return shared.ServiceInstance{
		Name:      name,
		Namespace: TestNamespace,
		Labels:    []string{"test", "test2"},
		PlanSpec: map[string]interface{}{
			"first": "1",
			"second": map[string]interface{}{
				"value": "2",
			},
		},
		ServicePlan: shared.ServicePlan{
			Name:         "86064792-7ea2-467b-af93-ac9694d96d52",
			ExternalName: "default",
		},
		ServiceClass: shared.ServiceClass{
			Name:         "4f6e6cf6-ffdd-425f-a2c7-3c9258ad2468",
			ExternalName: "user-provided-service",
			Namespace:    TestNamespace,
		},
		Status: shared.ServiceInstanceStatus{
			Type: shared.ServiceInstanceStatusTypeRunning,
		},
		Bindable: true,
	}
}

func instanceEvent(eventType string, serviceInstance shared.ServiceInstance) ServiceInstanceEvent {
	return ServiceInstanceEvent{
		Type:            eventType,
		ServiceInstance: serviceInstance,
	}
}

func readInstanceEvent(sub *graphql.Subscription) (ServiceInstanceEvent, error) {
	type Response struct {
		ServiceInstanceEvent ServiceInstanceEvent
	}
	var bindingEvent Response
	err := sub.Next(&bindingEvent, tester.DefaultSubscriptionTimeout)

	return bindingEvent.ServiceInstanceEvent, err
}

func checkInstanceEvent(t *testing.T, expected, actual ServiceInstanceEvent) {
	assert.Equal(t, expected.Type, actual.Type)
	assert.Equal(t, expected.ServiceInstance.Name, actual.ServiceInstance.Name)
}
