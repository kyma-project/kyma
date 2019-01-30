// +build acceptance

package k8s

import (
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"

	jsonEncoder "encoding/json"

	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/internal/client"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/internal/dex"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/internal/graphql"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/internal/waiter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	podName      = "test-pod"
	podNamespace = "ui-api-acceptance-pod"
)

type podQueryResponse struct {
	Pod pod `json:"pod"`
}

type podsQueryResponse struct {
	Pods []pod `json:"pods"`
}

type updatePodMutationResponse struct {
	UpdatePod pod `json:"updatePod"`
}

type deletePodMutationResponse struct {
	DeletePod pod `json:"deletePod"`
}

type pod struct {
	Name              string           `json:"name"`
	NodeName          string           `json:"nodeName"`
	Namespace         string           `json:"namespace"`
	RestartCount      int              `json:"restartCount"`
	CreationTimestamp int64            `json:"creationTimestamp"`
	Labels            labels           `json:"labels"`
	Status            podStatusType    `json:"status"`
	ContainerStates   []containerState `json:"containerStates"`
	JSON              json             `json:"json"`
}

type containerState struct {
	State   containerStateType `json:"state"`
	Reason  string             `json:"reason"`
	Message string             `json:"message"`
}

type containerStateType string

const (
	containerStateTypeWaiting    containerStateType = "WAITING"
	containerStateTypeRunning    containerStateType = "RUNNING"
	containerStateTypeTerminated containerStateType = "TERMINATED"
)

type labels map[string]string

type podStatusType string

const (
	podStatusTypePending   podStatusType = "PENDING"
	podStatusTypeRunning   podStatusType = "RUNNING"
	podStatusTypeSucceeded podStatusType = "SUCCEEDED"
	podStatusTypeFailed    podStatusType = "FAILED"
	podStatusTypeUnknown   podStatusType = "UNKNOWN"
)

type json map[string]interface{}

func TestPodQueries(t *testing.T) {
	dex.SkipTestIfSCIEnabled(t)

	c, err := graphql.New()
	require.NoError(t, err)

	k8sClient, _, err := client.NewClientWithConfig()
	require.NoError(t, err)

	t.Log("Creating namespace...")
	_, err = k8sClient.Namespaces().Create(fixNamespace(podNamespace))
	require.NoError(t, err)

	defer func() {
		t.Log("Deleting namespace...")
		err = k8sClient.Namespaces().Delete(podNamespace, &metav1.DeleteOptions{})
		require.NoError(t, err)
	}()

	t.Log("Creating pod...")
	_, err = k8sClient.Pods(podNamespace).Create(fixPod(podName, podNamespace))
	require.NoError(t, err)

	t.Log("Retrieving pod...")
	err = waiter.WaitAtMost(func() (bool, error) {
		_, err := k8sClient.Pods(podNamespace).Get(podName, metav1.GetOptions{})
		if err == nil {
			return true, nil
		}
		return false, err
	}, time.Minute)
	require.NoError(t, err)

	t.Log("Querying for pod...")
	var podRes podQueryResponse
	err = c.Do(fixPodQuery(), &podRes)
	require.NoError(t, err)
	assert.Equal(t, podName, podRes.Pod.Name)
	assert.Equal(t, podNamespace, podRes.Pod.Namespace)

	t.Log("Querying for pods...")
	var podsRes podsQueryResponse
	err = c.Do(fixPodsQuery(), &podsRes)
	require.NoError(t, err)
	assert.Equal(t, podName, podsRes.Pods[0].Name)
	assert.Equal(t, podNamespace, podsRes.Pods[0].Namespace)

	t.Log("Updating...")
	podRes.Pod.JSON["metadata"].(map[string]interface{})["labels"] = map[string]string{"foo": "bar"}
	update := stringifyJSON(podRes.Pod.JSON)
	var updateRes updatePodMutationResponse
	err = c.Do(fixUpdatePodMutation(update), &updateRes)
	require.NoError(t, err)
	assert.Equal(t, podName, updateRes.UpdatePod.Name)
	assert.Equal(t, podNamespace, updateRes.UpdatePod.Namespace)

	t.Log("Deleting pod...")
	var deleteRes deletePodMutationResponse
	err = c.Do(fixDeletePodMutation(), &deleteRes)
	require.NoError(t, err)
	assert.Equal(t, podName, deleteRes.DeletePod.Name)
	assert.Equal(t, podNamespace, deleteRes.DeletePod.Namespace)

	t.Log("Waiting for deletion...")
	err = waiter.WaitAtMost(func() (bool, error) {
		_, err := k8sClient.Pods(podNamespace).Get(podName, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			return true, nil
		}
		if err != nil {
			return false, err
		}
		return false, nil
	}, time.Minute)
	require.NoError(t, err)
}

func fixPod(name, namespace string) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  podName,
					Image: "invalid",
				},
			},
		},
	}
}

func fixPodQuery() *graphql.Request {
	query := `query ($name: String!, $namespace: String!) {
				pod(name: $name, namespace: $namespace) {
					name
					nodeName
					namespace
					restartCount
					creationTimestamp
					labels
					status
					containerStates {
						state
						reason
						message
					}
					json
				}
			}`
	req := graphql.NewRequest(query)
	req.SetVar("name", podName)
	req.SetVar("namespace", podNamespace)

	return req
}

func fixPodsQuery() *graphql.Request {
	query := `query ($namespace: String!) {
				pods(namespace: $namespace) {
					name
					nodeName
					namespace
					restartCount
					creationTimestamp
					labels
					status
					containerStates {
						state
						reason
						message
					}
					json
				}
			}`
	req := graphql.NewRequest(query)
	req.SetVar("namespace", podNamespace)

	return req
}

func fixUpdatePodMutation(pod string) *graphql.Request {
	mutation := `mutation ($name: String!, $namespace: String!, $pod: JSON!) {
					updatePod(name: $name, namespace: $namespace, pod: $pod) {
						name
						nodeName
						namespace
						restartCount
						creationTimestamp
						labels
						status
						containerStates {
							state
							reason
							message
						}
						json
					}
				}`
	req := graphql.NewRequest(mutation)
	req.SetVar("name", podName)
	req.SetVar("namespace", podNamespace)
	req.SetVar("pod", pod)

	return req
}

func fixDeletePodMutation() *graphql.Request {
	mutation := `mutation ($name: String!, $namespace: String!) {
					deletePod(name: $name, namespace: $namespace) {
						name
						nodeName
						namespace
						restartCount
						creationTimestamp
						labels
						status
						containerStates {
							state
							reason
							message
						}
						json
					}
				}`
	req := graphql.NewRequest(mutation)
	req.SetVar("name", podName)
	req.SetVar("namespace", podNamespace)

	return req
}

func stringifyJSON(in json) string {
	bytes, _ := jsonEncoder.Marshal(in)
	return string(bytes)
}
