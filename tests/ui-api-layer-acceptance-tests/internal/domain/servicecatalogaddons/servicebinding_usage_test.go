// +build acceptance

package servicecatalogaddons

import (
	"fmt"
	"testing"

	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/internal/domain/shared"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/internal/domain/shared/fixture"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/internal/domain/shared/wait"

	catalog "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	tester "github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/internal/client"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/internal/graphql"
)

type ServiceBindingUsageEvent struct {
	Type                string
	ServiceBindingUsage shared.ServiceBindingUsage
}

type DeleteServiceBindingUsageOutput struct {
	Name        string
	Environment string
}

type instanceQueryResponse struct {
	ServiceInstance shared.ServiceInstance
}

type bindingUsageQueryResponse struct {
	ServiceBindingUsage shared.ServiceBindingUsage
}

type bindingUsageCreateMutationResponse struct {
	CreateServiceBindingUsage shared.ServiceBindingUsage
}

type bindingUsageDeleteMutationResponse struct {
	DeleteServiceBindingUsage DeleteServiceBindingUsageOutput
}

func TestServiceBindingUsageMutationsAndQueries(t *testing.T) {
	// GIVEN
	suite := newBindingUsageSuite(t)
	suite.prepareInstanceAndBinding()
	defer suite.deleteServiceInstanceAndBinding()

	t.Log("Subscribe Binding Usage")
	subscription := suite.subscribeBindingUsage()
	defer subscription.Close()

	// WHEN
	t.Log("Create Binding Usage")
	createRes, err := suite.createBindingUsage()

	// THEN
	assert.NoError(t, err)
	suite.assertEqualBindingUsage(suite.givenBindingUsage, createRes.CreateServiceBindingUsage)

	// WHEN
	event, err := suite.readServiceBindingUsageEvent(subscription)

	// THEN
	t.Log("Check subscription event")
	assert.NoError(t, err)
	suite.assertEqualBindingUsageEvent(event)

	// WHEN
	t.Log("Query Single Binding Usage")
	res, err := suite.queryBindingUsage()

	// THEN
	assert.NoError(t, err)
	suite.assertEqualBindingUsage(suite.givenBindingUsage, res.ServiceBindingUsage)

	// WHEN
	t.Log("Query Binding Usage of Instance")
	instanceRes, err := suite.queryServiceInstance()

	// THEN
	assert.NoError(t, err)
	suite.assertServiceInstanceContainsServiceBindingUsage(instanceRes.ServiceInstance, suite.givenBindingUsage)

	// WHEN
	t.Log("Delete Binding Usage")
	deleteRes, err := suite.deleteBindingUsage()

	// THEN
	assert.NoError(t, err)
	suite.assertBindingUsageDeleteResponse(deleteRes)
}

func newBindingUsageSuite(t *testing.T) *bindingUsageTestSuite {
	c, err := graphql.New()
	require.NoError(t, err)
	svcatCli, _, err := client.NewServiceCatalogClientWithConfig()
	require.NoError(t, err)

	return &bindingUsageTestSuite{
		gqlCli:   c,
		svcatCli: svcatCli,
		t:        t,
	}
}

type bindingUsageTestSuite struct {
	gqlCli   *graphql.Client
	svcatCli *clientset.Clientset
	t        *testing.T

	givenBindingUsage shared.ServiceBindingUsage
	givenInstance     shared.ServiceInstance
	givenBinding      shared.ServiceBinding
}

func (s *bindingUsageTestSuite) fixServiceBindingUsage(name, serviceBindingName, deploymentName string) shared.ServiceBindingUsage {
	return shared.ServiceBindingUsage{
		Name:        name,
		Environment: TestNamespace,
		ServiceBinding: shared.ServiceBinding{
			Name:        serviceBindingName,
			Environment: TestNamespace,
		},
		UsedBy: shared.LocalObjectReference{
			Name: deploymentName,
			Kind: "deployment",
		},
	}
}

func (s *bindingUsageTestSuite) prepareInstanceAndBinding() {
	instanceName := "binding-usage-test"
	bindingName := "binding-usage-test"
	s.givenInstance = fixture.ServiceInstance(instanceName, TestNamespace)
	s.givenBinding = fixture.ServiceBinding(bindingName, instanceName, TestNamespace)
	s.givenBindingUsage = s.fixServiceBindingUsage("binding-usage-test", bindingName, "sample-deployment")

	s.t.Log("Create Instance")
	err := s.createInstance()
	require.NoError(s.t, err)

	s.t.Log("Wait for Instance")
	err = wait.ForServiceInstanceReady(s.givenInstance.Name, s.givenInstance.Environment, s.svcatCli)
	require.NoError(s.t, err)

	s.t.Log("Create Binding")
	err = s.createBinding()
	require.NoError(s.t, err)

	s.t.Log("Wait for Binding")
	err = wait.ForServiceBindingReady(s.givenBinding.Name, s.givenBinding.Environment, s.svcatCli)
	require.NoError(s.t, err)
}

func (s *bindingUsageTestSuite) createInstance() error {
	_, err := s.svcatCli.ServicecatalogV1beta1().ServiceInstances(TestNamespace).Create(&catalog.ServiceInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name: s.givenInstance.Name,
		},
		Spec: catalog.ServiceInstanceSpec{
			PlanReference: catalog.PlanReference{
				ClusterServiceClassExternalName: s.givenInstance.ClusterServiceClass.ExternalName,
				ClusterServicePlanExternalName:  s.givenInstance.ClusterServicePlan.ExternalName,
			},
		},
	})
	return err
}

func (s *bindingUsageTestSuite) createBinding() error {
	_, err := s.svcatCli.ServicecatalogV1beta1().ServiceBindings(TestNamespace).Create(&catalog.ServiceBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: s.givenBinding.Name,
		},
		Spec: catalog.ServiceBindingSpec{
			ServiceInstanceRef: catalog.LocalObjectReference{
				Name: s.givenInstance.Name,
			},
		},
	})
	return err
}

func (s *bindingUsageTestSuite) deleteInstance() error {
	return s.svcatCli.ServicecatalogV1beta1().ServiceInstances(TestNamespace).Delete(s.givenInstance.Name, &metav1.DeleteOptions{})
}

func (s *bindingUsageTestSuite) deleteBinding() error {
	return s.svcatCli.ServicecatalogV1beta1().ServiceBindings(TestNamespace).Delete(s.givenBinding.Name, &metav1.DeleteOptions{})
}

func (s *bindingUsageTestSuite) createBindingUsage() (bindingUsageCreateMutationResponse, error) {
	query := fmt.Sprintf(`
		mutation ($name: String!, $environment: String!, $serviceBindingRefName: String!, $usedByName: String!, $usedByKind: String!) {
			createServiceBindingUsage(createServiceBindingUsageInput: {
				name: $name,
				environment: $environment,
				serviceBindingRef: {
					name: $serviceBindingRefName,
				},
				usedBy: {
					name: $usedByName,
					kind: $usedByKind,
				},
			}) {
				%s
			}
		}
	`, s.bindingUsageDetailsFields())
	req := graphql.NewRequest(query)
	req.SetVar("name", s.givenBindingUsage.Name)
	req.SetVar("environment", s.givenBindingUsage.Environment)
	req.SetVar("serviceBindingRefName", s.givenBindingUsage.ServiceBinding.Name)
	req.SetVar("usedByName", s.givenBindingUsage.UsedBy.Name)
	req.SetVar("usedByKind", s.givenBindingUsage.UsedBy.Kind)

	var res bindingUsageCreateMutationResponse
	err := s.gqlCli.Do(req, &res)

	return res, err
}

func singleResourceQueryRequest(resourceDetailsQuery string, expectedResource shared.ServiceInstance) *graphql.Request {
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

func querySingleInstance(c *graphql.Client, resourceDetailsQuery string, expectedResource shared.ServiceInstance) (instanceQueryResponse, error) {
	req := singleResourceQueryRequest(resourceDetailsQuery, expectedResource)

	var res instanceQueryResponse
	err := c.Do(req, &res)

	return res, err
}

func (s *bindingUsageTestSuite) queryServiceInstance() (instanceQueryResponse, error) {
	return querySingleInstance(s.gqlCli, `
		name
		serviceBindingUsages {
			name
			environment
			serviceBinding {
				name
				serviceInstanceName
				environment
				secret {
					name
					environment
					data
				}
			}
			usedBy {
				kind
				name
			}
			status {
				type
			}
		}
	`, s.givenInstance)
}

func (s *bindingUsageTestSuite) deleteServiceInstanceAndBinding() {
	s.t.Log("Delete Binding")
	err := s.deleteBinding()
	assert.NoError(s.t, err)

	s.t.Log("Wait for binding deletion")
	err = wait.ForServiceBindingDeletion(s.givenBinding.Name, s.givenBinding.Environment, s.svcatCli)
	assert.NoError(s.t, err)

	s.t.Log("Delete Instance")
	err = s.deleteInstance()
	assert.NoError(s.t, err)

	s.t.Log("Wait for instance deletion")
	err = wait.ForServiceInstanceDeletion(s.givenBinding.Name, s.givenBinding.Environment, s.svcatCli)
	assert.NoError(s.t, err)
}

func (s *bindingUsageTestSuite) deleteBindingUsage() (bindingUsageDeleteMutationResponse, error) {
	query := `
		mutation ($name: String!, $environment: String!) {
			deleteServiceBindingUsage(serviceBindingUsageName: $name, environment: $environment) {
				name
				environment
			}
		}
	`
	req := graphql.NewRequest(query)
	req.SetVar("name", s.givenBindingUsage.Name)
	req.SetVar("environment", s.givenBindingUsage.Environment)

	var res bindingUsageDeleteMutationResponse
	err := s.gqlCli.Do(req, &res)

	return res, err
}

func (s *bindingUsageTestSuite) assertEqualBindingUsage(expected, actual shared.ServiceBindingUsage) {
	assert.Equal(s.t, expected.Name, actual.Name)
	assert.Equal(s.t, expected.Environment, actual.Environment)
	assert.Equal(s.t, expected.Name, actual.ServiceBinding.Name)
	assert.Equal(s.t, expected.Environment, actual.ServiceBinding.Environment)
	assert.Equal(s.t, expected.UsedBy.Name, actual.UsedBy.Name)
	assert.Equal(s.t, expected.UsedBy.Kind, actual.UsedBy.Kind)

	// The test is checking, if the status is retrieved without any error.
	// Does not matter, if it is READY or PENDING
	assert.NotEmpty(s.t, actual.Status)
	assert.NotEqual(s.t, shared.ServiceBindingUsageStatusTypeUnknown, actual.Status)
}

func (s *bindingUsageTestSuite) assertServiceInstanceContainsServiceBindingUsage(instance shared.ServiceInstance, expected shared.ServiceBindingUsage) {
	// check, if service instance contains expected binding usage
	assert.Condition(s.t, func() bool {
		for _, bu := range instance.ServiceBindingUsages {
			if bu.Name == s.givenBindingUsage.Name {
				s.assertEqualBindingUsage(expected, bu)
				return true
			}
		}

		return false
	}, "Resource does not exist")
}

func (s *bindingUsageTestSuite) queryBindingUsage() (bindingUsageQueryResponse, error) {
	query := fmt.Sprintf(`
		query ($name: String!, $environment: String!) {
			serviceBindingUsage(name: $name, environment: $environment) {
				%s
			}
		}
	`, s.bindingUsageDetailsFields())
	req := graphql.NewRequest(query)
	req.SetVar("name", s.givenBindingUsage.Name)
	req.SetVar("environment", s.givenBindingUsage.Environment)

	var res bindingUsageQueryResponse
	err := s.gqlCli.Do(req, &res)

	return res, err
}

func (s *bindingUsageTestSuite) assertBindingUsageDeleteResponse(response bindingUsageDeleteMutationResponse) {
	assert.Equal(s.t, s.givenBindingUsage.Name, response.DeleteServiceBindingUsage.Name)
	assert.Equal(s.t, s.givenBindingUsage.Environment, response.DeleteServiceBindingUsage.Environment)
}

func (s *bindingUsageTestSuite) bindingUsageDetailsFields() string {
	return `
		name
		environment
		serviceBinding {
			name
			serviceInstanceName
			environment
			secret {
				name
				environment
				data
			}
		}
		usedBy {
			kind
			name
		}
		status {
			type
		}
	`
}

func (s *bindingUsageTestSuite) subscribeBindingUsage() *graphql.Subscription {
	query := fmt.Sprintf(`
			subscription ($environment: String!) {
				serviceBindingUsageEvent(environment: $environment) {
					%s
				}
			}
		`, s.bindingUsageEventDetailsFields())
	req := graphql.NewRequest(query)
	req.SetVar("environment", s.givenBindingUsage.Environment)

	return s.gqlCli.Subscribe(req)
}

func (s *bindingUsageTestSuite) readServiceBindingUsageEvent(sub *graphql.Subscription) (ServiceBindingUsageEvent, error) {
	type Response struct {
		ServiceBindingUsageEvent ServiceBindingUsageEvent
	}
	var bindingEvent Response
	err := sub.Next(&bindingEvent, tester.DefaultSubscriptionTimeout)

	return bindingEvent.ServiceBindingUsageEvent, err
}

func (s *bindingUsageTestSuite) bindingUsageEventDetailsFields() string {
	return `
        type
        serviceBindingUsage {
			name
        }
    `
}

func (s *bindingUsageTestSuite) assertEqualBindingUsageEvent(event ServiceBindingUsageEvent) {
	assert.Equal(s.t, "ADD", event.Type)
	assert.Equal(s.t, s.givenBindingUsage.Name, event.ServiceBindingUsage.Name)
}
