package integration

import (
	"github.com/avast/retry-go"
	"github.com/kyma-project/kyma/components/console-backend-service/integration/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"testing"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func TestNamespaces(t *testing.T) {
	suite := givenNewTestNamespaceSuite(t, restConfig)

	err := givenUserCanAccessResource("", "namespaces", []string{"create", "get", "delete", "update"})
	require.NoError(t, err)

	t.Log("Creating namespace...")
	createRsp, err := suite.whenNamespaceIsCreated()
	suite.thenThereIsNoError(err)
	suite.thenThereIsNoGqlError(createRsp.GqlErrors)
	suite.thenCreateNamespaceResponseIsAsExpected(createRsp)
	suite.thenNamespaceExistsInK8s()

	t.Log("Querying for namespace...")
	queryRsp, err := suite.whenNamespaceIsQueried()
	suite.thenThereIsNoError(err)
	suite.thenThereIsNoGqlError(queryRsp.GqlErrors)
	suite.thenNamespaceResponseIsAsExpected(queryRsp)

	t.Log("Updating namespace...")
	updateResp, err := suite.whenNamespaceIsUpdated()
	suite.thenThereIsNoError(err)
	suite.thenThereIsNoGqlError(updateResp.GqlErrors)
	suite.thenUpdateNamespaceResponseIsAsExpected(updateResp)
	suite.thenNamespaceIsUpdatedInK8s()

	t.Log("Deleting namespace...")
	deleteRsp, err := suite.whenNamespaceIsDeleted()
	suite.thenThereIsNoError(err)
	suite.thenThereIsNoGqlError(deleteRsp.GqlErrors)
	suite.thenDeleteNamespaceResponseIsAsExpected(deleteRsp)
	suite.thenNamespaceIsTerminating()
}

func TestNamespacesForbidden(t *testing.T) {
	suite := givenNewTestNamespaceSuite(t, restConfig)

	err := givenUserCannotAccessResource("", "namespaces")
	require.NoError(t, err)

	thenRequestsShouldBeDenied(t, suite.gqlClient,
		suite.fixNamespaceCreate(),
		suite.fixNamespaceUpdate(),
		suite.fixNamespaceQuery(),
		suite.fixNamespaceDelete(),
	)
}

type testNamespaceSuite struct {
	t             *testing.T
	k8sClient     *corev1.CoreV1Client
	gqlClient     *graphql.Client
	namespaceName string
	labels        map[string]string
	updatedLabels map[string]string
}

func givenNewTestNamespaceSuite(t *testing.T, restConfig *rest.Config) testNamespaceSuite {
	k8s, err := corev1.NewForConfig(restConfig)
	require.NoError(t, err)

	return testNamespaceSuite{
		t:             t,
		k8sClient:     k8s,
		gqlClient:     graphql.New(gqlEndpoint),
		namespaceName: "test-namespace",
		labels: map[string]string{
			"aaa": "bbb",
		},
		updatedLabels: map[string]string{
			"ccc": "ddd",
		},
	}
}

func (s testNamespaceSuite) whenNamespaceIsCreated() (createNamespaceResponse, error) {
	var rsp createNamespaceResponse
	err := s.gqlClient.Do(s.fixNamespaceCreate(), &rsp)
	return rsp, err
}

func (s testNamespaceSuite) thenThereIsNoError(err error) {
	require.NoError(s.t, err)
}

func (s testNamespaceSuite) thenThereIsNoGqlError(gqlErr GqlErrors) {
	require.Empty(s.t, gqlErr.Errors)
}

func (s testNamespaceSuite) thenCreateNamespaceResponseIsAsExpected(rsp createNamespaceResponse) {
	assert.Equal(s.t, s.fixCreateNamespaceResponse(), rsp)
}

func (s testNamespaceSuite) thenNamespaceExistsInK8s() {
	ns, err := s.k8sClient.Namespaces().Get(s.namespaceName, metav1.GetOptions{})
	require.NoError(s.t, err)
	assert.Equal(s.t, ns.Name, s.namespaceName)
	assert.Equal(s.t, ns.Labels, s.labels)
}

func (s testNamespaceSuite) thenNamespaceIsUpdatedInK8s() {
	ns, err := s.k8sClient.Namespaces().Get(s.namespaceName, metav1.GetOptions{})
	require.NoError(s.t, err)
	assert.Equal(s.t, ns.Name, s.namespaceName)
	assert.Equal(s.t, ns.Labels, s.updatedLabels)
}

func (s testNamespaceSuite) whenNamespaceIsQueried() (namespaceResponse, error) {
	var rsp namespaceResponse
	err := s.gqlClient.Do(s.fixNamespaceQuery(), &rsp)
	return rsp, err
}

func (s testNamespaceSuite) thenNamespaceResponseIsAsExpected(rsp namespaceResponse) {
	assert.Equal(s.t, s.fixNamespaceResponse(), rsp)
}

func (s testNamespaceSuite) whenNamespaceIsUpdated() (updateNamespaceResponse, error) {
	var rsp updateNamespaceResponse
	err := s.gqlClient.Do(s.fixNamespaceUpdate(), &rsp)
	return rsp, err
}

func (s testNamespaceSuite) thenUpdateNamespaceResponseIsAsExpected(rsp updateNamespaceResponse) {
	assert.Equal(s.t, s.fixUpdateNamespaceResponse(), rsp)
}

func (s testNamespaceSuite) whenNamespaceIsDeleted() (deleteNamespaceResponse, error) {
	var rsp deleteNamespaceResponse
	err := s.gqlClient.Do(s.fixNamespaceDelete(), &rsp)
	return rsp, err
}

func (s testNamespaceSuite) thenDeleteNamespaceResponseIsAsExpected(rsp deleteNamespaceResponse) {
	assert.Equal(s.t, s.fixDeleteNamespaceResponse(), rsp)
}

func (s testNamespaceSuite) thenNamespaceIsTerminating() {
	err := retry.Do(func() error {
		ns, err := s.k8sClient.Namespaces().Get(s.namespaceName, metav1.GetOptions{})
		if apierrors.IsNotFound(err) || ns.Status.Phase == v1.NamespaceTerminating {
			return nil
		}

		return err
	})
	require.NoError(s.t, err)
}

func (s testNamespaceSuite) fixNamespaceObj() namespaceObj {
	return namespaceObj{
		Name:   s.namespaceName,
		Labels: s.labels,
	}
}

func (s testNamespaceSuite) fixNamespaceObjAfterUpdate() namespaceObj {
	return namespaceObj{
		Name:   s.namespaceName,
		Labels: s.updatedLabels,
	}
}

func (s testNamespaceSuite) fixCreateNamespaceResponse() createNamespaceResponse {
	return createNamespaceResponse{CreateNamespace: s.fixNamespaceObj()}
}

func (s testNamespaceSuite) fixNamespaceResponse() namespaceResponse {
	return namespaceResponse{Namespace: s.fixNamespaceObj()}
}

func (s testNamespaceSuite) fixUpdateNamespaceResponse() updateNamespaceResponse {
	return updateNamespaceResponse{UpdateNamespace: s.fixNamespaceObjAfterUpdate()}
}

func (s testNamespaceSuite) fixDeleteNamespaceResponse() deleteNamespaceResponse {
	return deleteNamespaceResponse{DeleteNamespace: s.fixNamespaceObjAfterUpdate()}
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
					labels
				  }
				}`
	req := graphql.NewRequest(query)
	req.SetVar("name", s.namespaceName)
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

type namespaceObj struct {
	Name   string `json:"name"`
	Labels labels `json:"labels"`
}

type GqlErrors struct {
	Errors []interface{} `json:"errors"`
}

type createNamespaceResponse struct {
	GqlErrors
	CreateNamespace namespaceObj `json:"createNamespace"`
}

type namespaceResponse struct {
	GqlErrors
	Namespace namespaceObj `json:"namespace"`
}

type updateNamespaceResponse struct {
	GqlErrors
	UpdateNamespace namespaceObj `json:"updateNamespace"`
}

type deleteNamespaceResponse struct {
	GqlErrors
	DeleteNamespace namespaceObj `json:"deleteNamespace"`
}

type labels map[string]string
