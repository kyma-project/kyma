// +build acceptance

package k8s

import (
	"math/rand"
	"testing"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/avast/retry-go"
	tester "github.com/kyma-project/kyma/tests/console-backend-service"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/client"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/graphql"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNamespace(t *testing.T) {
	suite := givenNewTestNamespaceSuite(t)

	t.Log("Subscribing to namespaces...")
	subscription := suite.whenNamespacesAreSubscribedTo()
	defer subscription.Close()

	t.Log("Creating namespace...")
	createRsp, err := suite.whenNamespaceIsCreated()
	suite.thenThereIsNoError(t, err)
	suite.thenThereIsNoGqlError(t, createRsp.GqlErrors)
	suite.thenCreateNamespaceResponseIsAsExpected(t, createRsp)
	suite.thenNamespaceExistsInK8s(t)
	suite.thenAddEventIsSent(t, subscription)

	t.Log("Quering for namespace...")
	queryRsp, err := suite.whenNamespaceIsQueried()
	suite.thenThereIsNoError(t, err)
	suite.thenThereIsNoGqlError(t, queryRsp.GqlErrors)
	suite.thenNamespaceResponseIsAsExpected(t, queryRsp)

	t.Log("Quering for namespaces...")
	listQueryRsp, err := suite.whenNamespacesAreQueried()
	suite.thenThereIsNoError(t, err)
	suite.thenThereIsNoGqlError(t, listQueryRsp.GqlErrors)
	suite.thenNamespacesResponseIsAsExpected(t, listQueryRsp, queryRsp)

	t.Log("Updating namespace...")
	updateResp, err := suite.whenNamespaceIsUpdated()
	suite.thenThereIsNoError(t, err)
	suite.thenThereIsNoGqlError(t, updateResp.GqlErrors)
	suite.thenUpdateNamespaceResponseIsAsExpected(t, updateResp)
	suite.thenNamespaceAfterUpdateExistsInK8s(t)
	suite.thenUpdateEventIsSent(t, subscription, "Active")

	t.Log("Adding pod to namespace...")
	err = suite.whenPodIsAdded(t)
	suite.thenThereIsNoError(t, err)
	suite.thenUpdateEventIsSent(t, subscription, "Active")

	t.Log("Deleting namespace...")
	deleteRsp, err := suite.whenNamespaceIsDeleted()
	suite.thenThereIsNoError(t, err)
	suite.thenThereIsNoGqlError(t, deleteRsp.GqlErrors)
	suite.thenDeleteNamespaceResponseIsAsExpected(t, deleteRsp)
	suite.thenNamespaceIsRemovedFromK8sEventually(t)
	//namespace changes its status to 'Terminating' first - delete event is sent after few seconds
	suite.thenUpdateEventIsSent(t, subscription, "Terminating")
}

type testNamespaceSuite struct {
	gqlClient     *graphql.Client
	k8sClient     *corev1.CoreV1Client
	namespaceName string
	labels        map[string]string
	updatedLabels map[string]string
}

func randomString() string {
	rand.Seed(time.Now().UnixNano())
	letterAndNumbersRunes := []rune("abcdefghijklmnopqrstuvwxyz0123456789")

	b := make([]rune, 5)
	for i := range b {
		b[i] = letterAndNumbersRunes[rand.Intn(len(letterAndNumbersRunes))]
	}
	return string(b)
}

func givenNewTestNamespaceSuite(t *testing.T) testNamespaceSuite {
	c, err := graphql.New()
	require.NoError(t, err)

	k8s, _, err := client.NewClientWithConfig()
	require.NoError(t, err)

	suite := testNamespaceSuite{
		gqlClient:     c,
		k8sClient:     k8s,
		namespaceName: "test-namespace-" + randomString(),
		labels: map[string]string{
			"aaa": "bbb",
		},
		updatedLabels: map[string]string{
			"ccc": "ddd",
		},
	}
	return suite
}

//subscribe
func (s testNamespaceSuite) whenNamespacesAreSubscribedTo() *graphql.Subscription {
	subscription := s.gqlClient.Subscribe(s.fixNamespacesSubscription())
	return subscription
}

//create
func (s testNamespaceSuite) whenNamespaceIsCreated() (createNamespaceResponse, error) {
	var rsp createNamespaceResponse
	err := s.gqlClient.Do(s.fixNamespaceCreate(), &rsp)
	return rsp, err
}

func (s testNamespaceSuite) thenCreateNamespaceResponseIsAsExpected(t *testing.T, rsp createNamespaceResponse) {
	assert.Equal(t, s.fixCreateNamespaceResponse(), rsp)
}

func (s testNamespaceSuite) thenNamespaceExistsInK8s(t *testing.T) {
	ns, err := s.k8sClient.Namespaces().Get(s.namespaceName, metav1.GetOptions{})
	require.NoError(t, err)
	assert.Equal(t, ns.Name, s.namespaceName)
	assert.Equal(t, ns.Labels, s.labels)
}

func (s testNamespaceSuite) thenAddEventIsSent(t *testing.T, subscription *graphql.Subscription) {
	expectedEvent := fixNamespaceEvent("ADD", namespaceObj{Name: s.namespaceName, IsSystemNamespace: false, Labels: s.labels, Status: "Active"})
	checkNamespaceEvent(t, expectedEvent, subscription)
}

//find
func (s testNamespaceSuite) whenNamespaceIsQueried() (namespaceResponse, error) {
	var rsp namespaceResponse
	err := s.gqlClient.Do(s.fixNamespaceQuery(), &rsp)
	return rsp, err
}

func (s testNamespaceSuite) thenNamespaceResponseIsAsExpected(t *testing.T, rsp namespaceResponse) {
	assert.Equal(t, s.fixNamespaceResponse(), rsp)
}

//list
func (s testNamespaceSuite) whenNamespacesAreQueried() (namespacesResponse, error) {
	var rsp namespacesResponse
	err := s.gqlClient.Do(s.fixNamespacesQuery(), &rsp)
	return rsp, err
}

func (s testNamespaceSuite) thenNamespacesResponseIsAsExpected(t *testing.T, namespacesList namespacesResponse, namespace namespaceResponse) {
	assert.Contains(t, namespacesList.Namespaces, namespace.Namespace)
}

//update
func (s testNamespaceSuite) whenNamespaceIsUpdated() (updateNamespaceResponse, error) {
	var rsp updateNamespaceResponse
	err := s.gqlClient.Do(s.fixNamespaceUpdate(), &rsp)
	return rsp, err
}

func (s testNamespaceSuite) thenUpdateNamespaceResponseIsAsExpected(t *testing.T, rsp updateNamespaceResponse) {
	assert.Equal(t, s.fixUpdateNamespaceResponse(), rsp)
}

func (s testNamespaceSuite) thenNamespaceAfterUpdateExistsInK8s(t *testing.T) {
	ns, err := s.k8sClient.Namespaces().Get(s.namespaceName, metav1.GetOptions{})
	require.NoError(t, err)
	assert.Equal(t, ns.Name, s.namespaceName)
	assert.Equal(t, ns.Labels, s.updatedLabels)
}

func (s testNamespaceSuite) thenUpdateEventIsSent(t *testing.T, subscription *graphql.Subscription, status string) {
	expectedEvent := fixNamespaceEvent("UPDATE", namespaceObj{Name: s.namespaceName, IsSystemNamespace: false, Labels: s.updatedLabels, Status: status})
	checkNamespaceEvent(t, expectedEvent, subscription)
}

func (s testNamespaceSuite) whenPodIsAdded(t *testing.T) error {
	k8sClient, _, err := client.NewClientWithConfig()
	require.NoError(t, err)
	_, err = k8sClient.Pods(s.namespaceName).Create(fixPod("test-pod", s.namespaceName))
	return err
}

//delete
func (s testNamespaceSuite) whenNamespaceIsDeleted() (deleteNamespaceResponse, error) {
	var rsp deleteNamespaceResponse
	err := s.gqlClient.Do(s.fixNamespaceDelete(), &rsp)
	return rsp, err
}

func (s testNamespaceSuite) thenDeleteNamespaceResponseIsAsExpected(t *testing.T, rsp deleteNamespaceResponse) {
	assert.Equal(t, s.fixDeleteNamespaceResponse(), rsp)
}

func (s testNamespaceSuite) thenNamespaceIsRemovedFromK8sEventually(t *testing.T) {
	ns, err := s.k8sClient.Namespaces().Get(s.namespaceName, metav1.GetOptions{})
	if !apierrors.IsNotFound(err) {
		require.NoError(t, err)
		assert.Equal(t, "Terminating", string(ns.Status.Phase))
	}
}

//errors
func (s testNamespaceSuite) thenThereIsNoError(t *testing.T, err error) {
	require.NoError(t, err)
}

func (s testNamespaceSuite) thenThereIsNoGqlError(t *testing.T, gqlErr GqlErrors) {
	require.Empty(t, gqlErr.Errors)
}

//helpers
func (s testNamespaceSuite) fixNamespaceObj() namespaceObj {
	return namespaceObj{
		Name:              s.namespaceName,
		IsSystemNamespace: false,
		Labels:            s.labels,
		Status:            "Active",
	}
}

func (s testNamespaceSuite) fixNamespaceMutationObj(labels map[string]string) namespaceMutationObj {
	return namespaceMutationObj{
		Name:   s.namespaceName,
		Labels: labels,
	}
}

func fixNamespaceEvent(eventType string, namespace namespaceObj) NamespaceEventObj {
	return NamespaceEventObj{
		Type:      eventType,
		Namespace: namespace,
	}
}

func (s testNamespaceSuite) fixCreateNamespaceResponse() createNamespaceResponse {
	return createNamespaceResponse{CreateNamespace: s.fixNamespaceMutationObj(s.labels)}
}

func (s testNamespaceSuite) fixNamespaceResponse() namespaceResponse {
	return namespaceResponse{Namespace: s.fixNamespaceObj()}
}

func (s testNamespaceSuite) fixUpdateNamespaceResponse() updateNamespaceResponse {
	return updateNamespaceResponse{UpdateNamespace: s.fixNamespaceMutationObj(s.updatedLabels)}
}

func (s testNamespaceSuite) fixDeleteNamespaceResponse() deleteNamespaceResponse {
	return deleteNamespaceResponse{DeleteNamespace: s.fixNamespaceMutationObj(s.updatedLabels)}
}

func readNamespaceEvent(subscription *graphql.Subscription) (NamespaceEventObj, error) {
	type Response struct {
		NamespaceEvent NamespaceEventObj
	}
	var namespaceEvent Response
	err := subscription.Next(&namespaceEvent, tester.DefaultSubscriptionTimeout)
	return namespaceEvent.NamespaceEvent, err
}

func checkNamespaceEvent(t *testing.T, expected NamespaceEventObj, subscription *graphql.Subscription) {
	err := retry.Do(func() error {
		event, err := readNamespaceEvent(subscription)
		if err != nil {
			return err
		}
		if !assert.ObjectsAreEqual(expected, event) {
			return errors.Errorf("unexpected event %#v", event)
		}
		return nil
	})
	require.NoError(t, err)
}

//queries
func (s testNamespaceSuite) fixNamespacesSubscription() *graphql.Request {
	sub := `subscription {
		namespaceEvent {
			type
			namespace {
				name
				isSystemNamespace
				labels
				status
			}
		}
	}`
	req := graphql.NewRequest(sub)
	return req
}

func (s testNamespaceSuite) fixNamespaceCreate() *graphql.Request {
	query := `mutation ($name: String!, $labels: Labels!) {
		  createNamespace(name: $name, labels: $labels) {
				name
				labels
		  }
	}`
	req := graphql.NewRequest(query)
	req.SetVar("name", s.namespaceName)
	req.SetVar("labels", s.labels)
	return req
}

func (s testNamespaceSuite) fixNamespaceQuery() *graphql.Request {
	query := `query ($name: String!) {
		  namespace(name: $name) {
				name
				isSystemNamespace
				labels
				status
		  }
	}`
	req := graphql.NewRequest(query)
	req.SetVar("name", s.namespaceName)
	return req
}

func (s testNamespaceSuite) fixNamespacesQuery() *graphql.Request {
	query := `query {
		  namespaces {
				name
				isSystemNamespace
				labels
				status
		  }
	}`
	req := graphql.NewRequest(query)
	return req
}

func (s testNamespaceSuite) fixNamespaceUpdate() *graphql.Request {
	query := `mutation ($name: String!, $labels: Labels!) {
		  updateNamespace(name: $name, labels: $labels) {
				name
				labels
		  }
	}`
	req := graphql.NewRequest(query)
	req.SetVar("name", s.namespaceName)
	req.SetVar("labels", s.updatedLabels)
	return req
}

func (s testNamespaceSuite) fixNamespaceDelete() *graphql.Request {
	query := `mutation ($name: String!) {
		  deleteNamespace(name: $name) {
				name
				labels
		  }
	}`
	req := graphql.NewRequest(query)
	req.SetVar("name", s.namespaceName)
	return req
}

//types
type namespaceMutationObj struct {
	Name   string `json:"name"`
	Labels labels `json:"labels"`
}

type namespaceObj struct {
	Name              string `json:"name"`
	IsSystemNamespace bool   `json:"isSystemNamespace"`
	Labels            labels `json:"labels"`
	Status            string `json:"status"`
}

type NamespaceEventObj struct {
	Type      string
	Namespace namespaceObj
}

type GqlErrors struct {
	Errors []interface{} `json:"errors"`
}

type createNamespaceResponse struct {
	GqlErrors
	CreateNamespace namespaceMutationObj `json:"createNamespace"`
}

type namespaceResponse struct {
	GqlErrors
	Namespace namespaceObj `json:"namespace"`
}

type namespacesResponse struct {
	GqlErrors
	Namespaces []namespaceObj `json:"namespaces"`
}

type updateNamespaceResponse struct {
	GqlErrors
	UpdateNamespace namespaceMutationObj `json:"updateNamespace"`
}

type deleteNamespaceResponse struct {
	GqlErrors
	DeleteNamespace namespaceMutationObj `json:"deleteNamespace"`
}
