// +build acceptance

package k8s

import (
	"fmt"
	"testing"
	"time"

	"github.com/kyma-project/kyma/tests/console-backend-service/internal/client"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared/auth"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/graphql"
	"github.com/kyma-project/kyma/tests/console-backend-service/pkg/retrier"
	"github.com/kyma-project/kyma/tests/console-backend-service/pkg/waiter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	replicaSetName      = "test-replicaset"
)

type replicaSetQueryResponse struct {
	ReplicaSet replicaSet `json:"replicaSet"`
}

type replicaSetsQueryResponse struct {
	ReplicaSets []replicaSet `json:"replicaSets"`
}

type updateReplicaSetMutationResponse struct {
	UpdateReplicaSet replicaSet `json:"updateReplicaSet"`
}

type deleteReplicaSetMutationResponse struct {
	DeleteReplicaSet replicaSet `json:"deleteReplicaSet"`
}

type replicaSet struct {
	Name              string   `json:"name"`
	Namespace         string   `json:"namespace"`
	Pods              string   `json:"pods"`
	CreationTimestamp int64    `json:"creationTimestamp"`
	Images            []string `json:"images"`
	Labels            labels   `json:"labels"`
	JSON              json     `json:"json"`
}

func TestReplicaSet(t *testing.T) {
	c, err := graphql.New()
	require.NoError(t, err)

	appsV1Client, _, err := client.NewAppsClientWithConfig()
	require.NoError(t, err)

	t.Log("Creating replicaSet...")
	_, err = appsV1Client.ReplicaSets(testNamespace).Create(fixReplicaSet(replicaSetName, testNamespace))
	require.NoError(t, err)

	t.Log("Retrieving replicaSet...")
	err = waiter.WaitAtMost(func() (bool, error) {
		_, err := appsV1Client.ReplicaSets(testNamespace).Get(replicaSetName, metav1.GetOptions{})
		if err == nil {
			return true, nil
		}
		return false, err
	}, time.Minute)
	require.NoError(t, err)

	t.Log("Querying for replicaSet...")
	var replicaSetRes replicaSetQueryResponse
	err = c.Do(fixReplicaSetQuery(), &replicaSetRes)
	require.NoError(t, err)
	assert.Equal(t, replicaSetName, replicaSetRes.ReplicaSet.Name)
	assert.Equal(t, testNamespace, replicaSetRes.ReplicaSet.Namespace)

	t.Log("Querying for replicaSets...")
	var replicaSetsRes replicaSetsQueryResponse
	err = c.Do(fixReplicaSetsQuery(), &replicaSetsRes)
	require.NoError(t, err)
	assert.Equal(t, replicaSetName, replicaSetsRes.ReplicaSets[0].Name)
	assert.Equal(t, testNamespace, replicaSetsRes.ReplicaSets[0].Namespace)

	t.Log("Updating...")
	var updateRes updateReplicaSetMutationResponse
	err = retrier.Retry(func() error {
		err = c.Do(fixReplicaSetQuery(), &replicaSetRes)
		if err != nil {
			return err
		}
		replicaSetRes.ReplicaSet.JSON["metadata"].(map[string]interface{})["labels"] = map[string]string{"foo": "bar"}
		update, err := stringifyJSON(replicaSetRes.ReplicaSet.JSON)
		if err != nil {
			return err
		}
		err = c.Do(fixUpdateReplicaSetMutation(update), &updateRes)
		if err != nil {
			return err
		}
		return nil
	}, retrier.UpdateRetries)
	require.NoError(t, err)
	assert.Equal(t, replicaSetName, updateRes.UpdateReplicaSet.Name)
	assert.Equal(t, testNamespace, updateRes.UpdateReplicaSet.Namespace)

	t.Log("Deleting replicaSet...")
	var deleteRes deleteReplicaSetMutationResponse
	err = c.Do(fixDeleteReplicaSetMutation(), &deleteRes)
	require.NoError(t, err)
	assert.Equal(t, replicaSetName, deleteRes.DeleteReplicaSet.Name)
	assert.Equal(t, testNamespace, deleteRes.DeleteReplicaSet.Namespace)

	t.Log("Waiting for deletion...")
	err = waiter.WaitAtMost(func() (bool, error) {
		_, err := appsV1Client.ReplicaSets(testNamespace).Get(replicaSetName, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			return true, nil
		}
		if err != nil {
			return false, err
		}
		return false, nil
	}, time.Minute)
	require.NoError(t, err)

	t.Log("Checking authorization directives...")
	ops := &auth.OperationsInput{
		auth.Get:    {fixReplicaSetQuery()},
		auth.List:   {fixReplicaSetsQuery()},
		auth.Update: {fixUpdateReplicaSetMutation("{\"\":\"\"}")},
		auth.Delete: {fixDeleteReplicaSetMutation()},
	}
	AuthSuite.Run(t, ops)

}

func fixReplicaSet(name, namespace string) *appsv1.ReplicaSet {
	return &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1.ReplicaSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"foo": "bar",
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"foo": "bar",
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "foocontainer",
							Image: "invalid",
						},
					},
				},
			},
		},
	}
}

func getReplicaSetQueryAttributes() string {
	return `name
			namespace
			creationTimestamp
			labels
			pods
			images
			json`
}

func fixReplicaSetQuery() *graphql.Request {
	queryAttributes := getReplicaSetQueryAttributes()
	query := fmt.Sprintf(`query ($name: String!, $namespace: String!) {
				replicaSet(name: $name, namespace: $namespace) {
					%s
				}
			}`, queryAttributes)
	req := graphql.NewRequest(query)
	req.SetVar("name", replicaSetName)
	req.SetVar("namespace", testNamespace)

	return req
}

func fixReplicaSetsQuery() *graphql.Request {
	queryAttributes := getReplicaSetQueryAttributes()
	query := fmt.Sprintf(`query ($namespace: String!) {
				replicaSets(namespace: $namespace) {
					%s
				}
			}`, queryAttributes)
	req := graphql.NewRequest(query)
	req.SetVar("namespace", testNamespace)

	return req
}

func fixUpdateReplicaSetMutation(replicaSet string) *graphql.Request {
	mutation := `mutation ($name: String!, $namespace: String!, $replicaSet: JSON!) {
					updateReplicaSet(name: $name, namespace: $namespace, replicaSet: $replicaSet) {
						name
						namespace
						creationTimestamp
						labels
						pods
						images
						json
					}
				}`
	req := graphql.NewRequest(mutation)
	req.SetVar("name", replicaSetName)
	req.SetVar("namespace", testNamespace)
	req.SetVar("replicaSet", replicaSet)

	return req
}

func fixDeleteReplicaSetMutation() *graphql.Request {
	mutation := `mutation ($name: String!, $namespace: String!) {
					deleteReplicaSet(name: $name, namespace: $namespace) {
						name
						namespace
						creationTimestamp
						labels
						pods
						images
						json
					}
				}`
	req := graphql.NewRequest(mutation)
	req.SetVar("name", replicaSetName)
	req.SetVar("namespace", testNamespace)

	return req
}
