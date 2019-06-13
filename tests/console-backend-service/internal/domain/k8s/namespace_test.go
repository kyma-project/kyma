// +build acceptance

package k8s

import (
	"testing"

	"github.com/kyma-project/kyma/tests/console-backend-service/internal/client"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/dex"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNamespace(t *testing.T) {
	dex.SkipTestIfSCIEnabled(t)

	c, err := graphql.New()
	require.NoError(t, err)

	k8s, _, err := client.NewClientWithConfig()
	require.NoError(t, err)

	name := "tesr-namespace"
	labels := labels{
		"aaa": "bbb",
	}

	var rsp namespaceResponse
	err = c.Do(fixNamespaceCreate(name, labels), &rsp)
	require.NoError(t, err)
	assert.Equal(t, fixNamespaceResponse(name, labels), rsp)

	ns, err := k8s.Namespaces().Get(name, metav1.GetOptions{})
	require.NoError(t, err)
	assert.Equal(t, ns.Name, name)
	assert.Equal(t, ns.Labels, labels)

	rsp = namespaceResponse{}
	err = c.Do(fixNamespaceQuery(name), &rsp)
	require.NoError(t, err)
	assert.Equal(t, fixNamespaceResponse(name, labels), rsp)

	rsp = namespaceResponse{}
	err = c.Do(fixNamespaceDelete(name), &rsp)
	require.NoError(t, err)
	assert.Equal(t, fixNamespaceResponse(name, labels), rsp)

	_, err = k8s.Namespaces().Get(name, metav1.GetOptions{})
	require.Error(t, err)

}

func fixNamespaceResponse(name string, labels labels) namespaceResponse {
	return namespaceResponse{
		name:   name,
		labels: labels,
	}
}

func fixNamespaceCreate(name string, labels labels) *graphql.Request {
	query := `mutation ($name: String!, $labels: Labels!) {
				  createNamespace(name: $name, labels: $labels) {
					name
					labels
				  }
				}`
	req := graphql.NewRequest(query)
	req.SetVar("name", name)
	req.SetVar("labels", labels)
	return req
}

func fixNamespaceQuery(name string) *graphql.Request {
	query := `query ($name: String!) {
				  namespace(name: $name) {
					name
					labels
				  }
				}`
	req := graphql.NewRequest(query)
	req.SetVar("name", name)
	return req
}

func fixNamespaceDelete(name string) *graphql.Request {
	query := `mutation ($name: String!) {
				  deleteNamespace(name: $name) {
					name
					labels
				  }
				}`
	req := graphql.NewRequest(query)
	req.SetVar("name", name)
	return req
}

type namespaceResponse struct {
	name   string
	labels labels
}
