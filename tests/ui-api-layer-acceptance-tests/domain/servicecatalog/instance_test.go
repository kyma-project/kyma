// +build acceptance

package servicecatalog

import (
	"fmt"
	"testing"
	"time"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/dex"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/graphql"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/k8s"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/waiter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	instanceReadyTimeout    = time.Second * 45
	instanceDeletionTimeout = time.Second * 15
)

type ServiceInstanceEvent struct {
	Type     string
	Instance ServiceInstance
}

type ServiceInstance struct {
	Name                    string
	Environment             string
	ServiceClassName        string
	ServiceClassDisplayName string
	ServicePlanName         string
	ServicePlanDisplayName  string
	ServicePlanSpec         map[string]interface{}
	ServicePlan             ServicePlan
	ServiceClass            ServiceClass
	CreationTimestamp       int
	Labels                  []string
	Status                  ServiceInstanceStatus
	ServiceBindings         []ServiceBinding
	ServiceBindingUsages    []ServiceBindingUsage
}

type ServiceInstanceStatus struct {
	Type    string
	Reason  string
	Message string
}

type instancesQueryResponse struct {
	ServiceInstances []ServiceInstance
}

type instanceQueryResponse struct {
	ServiceInstance ServiceInstance
}

type instanceCreateMutationResponse struct {
	CreateServiceInstance ServiceInstance
}

type instanceDeleteMutationResponse struct {
	DeleteServiceInstance ServiceInstance
}

func TestInstanceMutationsAndQueries(t *testing.T) {
	if dex.IsSCIEnabled() {
		t.Skip("SCI Enabled")
	}

	c, err := graphql.New()
	require.NoError(t, err)

	svcatCli, _, err := k8s.NewServiceCatalogClientWithConfig()
	require.NoError(t, err)

	expectedResource := instance("test-instance")
	resourceDetailsQuery := instanceDetailsFields()

	t.Log("Subscribe instances")
	subscription := subscribeInstance(c, instanceEventDetailsFields(), expectedResource.Environment)
	defer subscription.Close()

	t.Log("Create instance")
	createRes, err := createInstance(c, resourceDetailsQuery, expectedResource)

	require.NoError(t, err)
	checkInstance(t, expectedResource, createRes.CreateServiceInstance)

	t.Log("Check subscription event")
	expectedEvent := instanceEvent("ADD", expectedResource)
	event, err := readInstanceEvent(subscription)
	assert.NoError(t, err)
	checkInstanceEvent(t, expectedEvent, event)

	t.Log("Wait For Instance Ready")
	err = waitForInstanceReady(expectedResource.Name, expectedResource.Environment, svcatCli)
	assert.NoError(t, err)

	t.Log("Query Single Resource")
	res, err := querySingleInstance(c, resourceDetailsQuery, expectedResource)

	assert.NoError(t, err)
	checkInstance(t, expectedResource, res.ServiceInstance)

	t.Log("Query Multiple Resources")
	multipleRes, err := queryMultipleInstances(c, resourceDetailsQuery, expectedResource)

	assert.NoError(t, err)
	assertInstanceExistsAndEqual(t, expectedResource, multipleRes.ServiceInstances)

	// We must again wait for RUNNING status of instance, because sometimes Kubernetess change status from RUNNING to PROVISIONING at the first query - Query Single Resource
	t.Log("Wait For Instance Ready")
	err = waitForInstanceReady(expectedResource.Name, expectedResource.Environment, svcatCli)
	assert.NoError(t, err)

	t.Log("Query Multiple Resources With Status")
	multipleResWithStatus, err := queryMultipleInstancesWithStatus(c, resourceDetailsQuery, expectedResource)

	assert.NoError(t, err)
	assertInstanceExistsAndEqual(t, expectedResource, multipleResWithStatus.ServiceInstances)

	t.Log("Delete instance")
	deleteRes, err := deleteInstance(c, resourceDetailsQuery, expectedResource)

	assert.NoError(t, err)
	checkInstance(t, expectedResource, deleteRes.DeleteServiceInstance)

	t.Log("Wait For Instance Deletion")
	err = waitForInstanceDeletion(expectedResource.Name, expectedResource.Environment, svcatCli)
	assert.NoError(t, err)
}

func createInstance(c *graphql.Client, resourceDetailsQuery string, expectedResource ServiceInstance) (instanceCreateMutationResponse, error) {
	query := fmt.Sprintf(`
			mutation ($name: String!, $environment: String!, $externalPlanName: String!, $externalServiceClassName: String!, $labels: [String!]!, $parameterSchema: JSON) {
				createServiceInstance(params:{
    				name: $name,
    				environment: $environment,
    				externalPlanName: $externalPlanName,
    				externalServiceClassName: $externalServiceClassName,
    				labels: $labels,
					parameterSchema: $parameterSchema
				}) {
					%s
				}
			}
		`, resourceDetailsQuery)
	req := graphql.NewRequest(query)
	req.SetVar("name", expectedResource.Name)
	req.SetVar("environment", expectedResource.Environment)
	req.SetVar("externalPlanName", expectedResource.ServicePlan.ExternalName)
	req.SetVar("externalServiceClassName", expectedResource.ServiceClass.ExternalName)
	req.SetVar("labels", expectedResource.Labels)
	req.SetVar("parameterSchema", expectedResource.ServicePlanSpec)

	var res instanceCreateMutationResponse
	err := c.Do(req, &res)

	return res, err
}

func subscribeInstance(c *graphql.Client, resourceDetailsQuery string, environment string) *graphql.Subscription {
	query := fmt.Sprintf(`
			subscription ($environment: String!) {
				serviceInstanceEvent(environment: $environment) {
					%s
				}
			}
		`, resourceDetailsQuery)
	req := graphql.NewRequest(query)
	req.SetVar("environment", environment)

	return c.Subscribe(req)
}

func querySingleInstance(c *graphql.Client, resourceDetailsQuery string, expectedResource ServiceInstance) (instanceQueryResponse, error) {
	req := singleResourceQueryRequest(resourceDetailsQuery, expectedResource)

	var res instanceQueryResponse
	err := c.Do(req, &res)

	return res, err
}

func queryMultipleInstances(c *graphql.Client, resourceDetailsQuery string, expectedResource ServiceInstance) (instancesQueryResponse, error) {
	query := fmt.Sprintf(`
			query ($environment: String!) {
				serviceInstances(environment: $environment) {
					%s
				}
			}	
		`, resourceDetailsQuery)
	req := graphql.NewRequest(query)
	req.SetVar("environment", expectedResource.Environment)

	var res instancesQueryResponse
	err := c.Do(req, &res)

	return res, err
}

func queryMultipleInstancesWithStatus(c *graphql.Client, resourceDetailsQuery string, expectedResource ServiceInstance) (instancesQueryResponse, error) {
	query := fmt.Sprintf(`
			query ($environment: String!, $status: InstanceStatusType) {
				serviceInstances(environment: $environment, status: $status) {
					%s
				}
			}	
		`, resourceDetailsQuery)
	req := graphql.NewRequest(query)
	req.SetVar("environment", expectedResource.Environment)
	req.SetVar("status", "RUNNING")

	var res instancesQueryResponse
	err := c.Do(req, &res)

	return res, err
}

func deleteInstance(c *graphql.Client, resourceDetailsQuery string, expectedResource ServiceInstance) (instanceDeleteMutationResponse, error) {
	query := fmt.Sprintf(`
			mutation ($name: String!, $environment: String!) {
				deleteServiceInstance(name: $name, environment: $environment) {
					%s
				}
			}
		`, resourceDetailsQuery)
	req := graphql.NewRequest(query)
	req.SetVar("name", expectedResource.Name)
	req.SetVar("environment", expectedResource.Environment)

	var res instanceDeleteMutationResponse
	err := c.Do(req, &res)

	return res, err
}

func singleResourceQueryRequest(resourceDetailsQuery string, expectedResource ServiceInstance) *graphql.Request {
	query := fmt.Sprintf(`
			query ($name: String!, $environment: String!) {
				serviceInstance(name: $name, environment: $environment) {
					%s
				}
			}
		`, resourceDetailsQuery)
	req := graphql.NewRequest(query)
	req.SetVar("name", expectedResource.Name)
	req.SetVar("environment", expectedResource.Environment)

	return req
}

func instanceDetailsFields() string {
	return `
		name
		environment
		serviceClassName
		ServiceClassDisplayName
		servicePlanName
		servicePlanDisplayName
		servicePlanSpec
		creationTimestamp
		labels
		status {
			type
			reason
			message
		}
		servicePlan {
			name
			displayName
			externalName
			description
			relatedServiceClassName
			instanceCreateParameterSchema
		}
		serviceClass {
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
		serviceBindings {
			name
			serviceInstanceName
			environment
			secret {
				name
				environment
				data
			}
		}
		serviceBindingUsages {
			name
			environment
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
        instance {
			%s
        }
    `, instanceDetailsFields())
}

func checkInstance(t *testing.T, expected, actual ServiceInstance) {
	assert.Equal(t, expected.Name, actual.Name)
	assert.Equal(t, expected.Environment, actual.Environment)
	assert.Equal(t, expected.ServicePlan.Name, actual.ServicePlan.Name)
	assert.Equal(t, expected.ServiceClass.Name, actual.ServiceClass.Name)
}

func assertInstanceExistsAndEqual(t *testing.T, expectedElement ServiceInstance, arr []ServiceInstance) {
	assert.Condition(t, func() (success bool) {
		for _, v := range arr {
			if v.Name == expectedElement.Name {
				checkInstance(t, expectedElement, v)
				return true
			}
		}

		return false
	}, "Resource does not exist")
}

func instance(name string) ServiceInstance {
	return ServiceInstance{
		Name:        name,
		Environment: tester.DefaultNamespace,
		Labels:      []string{"test", "test2"},
		ServicePlanSpec: map[string]interface{}{
			"first": "1",
			"second": map[string]interface{}{
				"value": "2",
			},
		},
		ServicePlan: ServicePlan{
			Name:         "86064792-7ea2-467b-af93-ac9694d96d52",
			ExternalName: "default",
		},
		ServiceClass: ServiceClass{
			Name:         "4f6e6cf6-ffdd-425f-a2c7-3c9258ad2468",
			ExternalName: "user-provided-service",
		},
	}
}

func instanceEvent(eventType string, serviceInstance ServiceInstance) ServiceInstanceEvent {
	return ServiceInstanceEvent{
		Type:     eventType,
		Instance: serviceInstance,
	}
}

func waitForInstanceReady(instanceName, environment string, svcatCli *clientset.Clientset) error {
	return waiter.WaitAtMost(func() (bool, error) {
		instance, err := svcatCli.ServicecatalogV1beta1().ServiceInstances(environment).Get(instanceName, metav1.GetOptions{})
		if err != nil || instance == nil {
			return false, err
		}

		conditions := instance.Status.Conditions
		for _, cond := range conditions {
			if cond.Type == v1beta1.ServiceInstanceConditionReady {
				return cond.Status == v1beta1.ConditionTrue, nil
			}
		}

		return false, nil
	}, instanceReadyTimeout)
}

func waitForInstanceDeletion(instanceName, environment string, svcatCli *clientset.Clientset) error {
	return waiter.WaitAtMost(func() (bool, error) {
		_, err := svcatCli.ServicecatalogV1beta1().ServiceInstances(environment).Get(instanceName, metav1.GetOptions{})

		if errors.IsNotFound(err) {
			return true, nil
		}
		if err != nil {
			return false, err
		}

		return false, nil
	}, instanceDeletionTimeout)
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
	assert.Equal(t, expected.Type, expected.Type)
	assert.Equal(t, expected.Instance.Name, expected.Instance.Name)
}
