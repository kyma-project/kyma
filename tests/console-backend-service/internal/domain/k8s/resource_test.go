// +build acceptance

package k8s

import (
	"testing"
	"time"

	"github.com/kyma-project/kyma/tests/console-backend-service/internal/client"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared/auth"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/graphql"
	"github.com/kyma-project/kyma/tests/console-backend-service/pkg/waiter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	resourceName       = "test-resource"
	resourceKind       = "Pod"
	resourceAPIVersion = "v1"
	resourceNamespace  = "console-backend-service-resource"
)

type createResourceMutationResponse struct {
	CreateResource json `json:"createResource"`
}

func TestResource(t *testing.T) {
	c, err := graphql.New()
	require.NoError(t, err)

	k8sClient, _, err := client.NewClientWithConfig()
	require.NoError(t, err)

	t.Log("Creating namespace...")
	_, err = k8sClient.Namespaces().Create(fixNamespace(resourceNamespace))
	require.NoError(t, err)

	defer func() {
		t.Log("Deleting namespace...")
		err = k8sClient.Namespaces().Delete(resourceNamespace, &metav1.DeleteOptions{})
		require.NoError(t, err)
	}()

	var createRes createResourceMutationResponse
	resourceJSON, err := fixResourceJSON()
	require.NoError(t, err)
	err = c.Do(fixCreateResourceMutation(resourceNamespace, resourceJSON), &createRes)
	require.NoError(t, err)
	require.NotNil(t, createRes)

	validateResource(t, createRes.CreateResource)

	t.Log("Retrieving resource...")
	err = waiter.WaitAtMost(func() (bool, error) {
		_, err := k8sClient.Pods(resourceNamespace).Get(resourceName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		return true, nil
	}, time.Minute)
	require.NoError(t, err)

	t.Log("Checking authorization directives...")
	ops := &auth.OperationsInput{
		auth.Create: {fixCreateResourceMutation(resourceNamespace, resourceJSON)},
	}
	AuthSuite.Run(t, ops)
}

func fixCreateResourceMutation(namespace, resource string) *graphql.Request {
	mutation := `mutation ($namespace: String!, $resource: JSON!) {
					createResource(namespace: $namespace, resource: $resource)
				}`
	req := graphql.NewRequest(mutation)
	req.SetVar("namespace", namespace)
	req.SetVar("resource", resource)

	return req
}

func fixResourceJSON() (string, error) {
	resource := map[string]interface{}{
		"kind":       resourceKind,
		"apiVersion": resourceAPIVersion,
		"metadata": map[string]interface{}{
			"name":      resourceName,
			"namespace": resourceNamespace,
		},
	}

	return stringifyJSON(resource)
}

func validateResource(t *testing.T, resource map[string]interface{}) {
	createdMeta, ok := resource["metadata"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, resourceNamespace, createdMeta["namespace"])
	assert.Equal(t, resourceName, createdMeta["name"])
	assert.Equal(t, resourceKind, resource["kind"])
	assert.Equal(t, resourceAPIVersion, resource["apiVersion"])
	assert.NotEmpty(t, createdMeta["creationTimestamp"])
}
