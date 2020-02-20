// +build acceptance

package k8s

import (
	"testing"
	"time"

	"github.com/kyma-project/kyma/tests/console-backend-service/pkg/retrier"

	tester "github.com/kyma-project/kyma/tests/console-backend-service"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/client"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared/auth"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/graphql"
	"github.com/kyma-project/kyma/tests/console-backend-service/pkg/waiter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	podName = "test-pod"
)

type PodEvent struct {
	Type string
	Pod  pod
}

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

func TestPod(t *testing.T) {
	c, err := graphql.New()
	require.NoError(t, err)

	k8sClient, _, err := client.NewClientWithConfig()
	require.NoError(t, err)

	t.Log("Subscribing to pods...")
	subscription := c.Subscribe(fixPodSubscription())
	defer subscription.Close()

	t.Log("Creating pod...")
	_, err = k8sClient.Pods(testNamespace).Create(fixPod(podName, testNamespace))
	require.NoError(t, err)

	t.Log("Retrieving pod...")
	err = waiter.WaitAtMost(func() (bool, error) {
		_, err := k8sClient.Pods(testNamespace).Get(podName, metav1.GetOptions{})
		if err == nil {
			return true, nil
		}
		return false, err
	}, time.Minute)
	require.NoError(t, err)

	t.Log("Checking subscription for created pod...")
	expectedEvent := podEvent("ADD", pod{Name: podName})
	assert.NoError(t, checkPodEvent(expectedEvent, subscription))

	t.Log("Querying for pod...")
	var podRes podQueryResponse
	err = c.Do(fixPodQuery(), &podRes)
	require.NoError(t, err)
	assert.Equal(t, podName, podRes.Pod.Name)
	assert.Equal(t, testNamespace, podRes.Pod.Namespace)

	t.Log("Querying for pods...")
	var podsRes podsQueryResponse
	err = c.Do(fixPodsQuery(), &podsRes)
	require.NoError(t, err)
	assert.Equal(t, podName, podsRes.Pods[0].Name)
	assert.Equal(t, testNamespace, podsRes.Pods[0].Namespace)

	t.Log("Updating...")
	var updateRes updatePodMutationResponse
	err = retrier.Retry(func() error {
		err = c.Do(fixPodQuery(), &podRes)
		if err != nil {
			return err
		}
		podRes.Pod.JSON["metadata"].(map[string]interface{})["labels"] = map[string]string{"foo": "bar"}
		update, err := stringifyJSON(podRes.Pod.JSON)
		if err != nil {
			return err
		}
		err = c.Do(fixUpdatePodMutation(update), &updateRes)
		if err != nil {
			return err
		}
		return nil
	}, retrier.UpdateRetries)
	require.NoError(t, err)
	assert.Equal(t, podName, updateRes.UpdatePod.Name)
	assert.Equal(t, testNamespace, updateRes.UpdatePod.Namespace)

	t.Log("Checking subscription for updated pod...")
	expectedEvent = podEvent("UPDATE", pod{Name: podName})
	assert.NoError(t, checkPodEvent(expectedEvent, subscription))

	t.Log("Deleting pod...")
	var deleteRes deletePodMutationResponse
	err = c.Do(fixDeletePodMutation(), &deleteRes)
	require.NoError(t, err)
	assert.Equal(t, podName, deleteRes.DeletePod.Name)
	assert.Equal(t, testNamespace, deleteRes.DeletePod.Namespace)

	t.Log("Waiting for deletion...")
	err = waiter.WaitAtMost(func() (bool, error) {
		_, err := k8sClient.Pods(testNamespace).Get(podName, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			return true, nil
		}
		if err != nil {
			return false, err
		}
		return false, nil
	}, time.Minute)
	require.NoError(t, err)

	t.Log("Checking subscription for deleted pod...")
	expectedEvent = podEvent("DELETE", pod{Name: podName})
	assert.NoError(t, checkPodEvent(expectedEvent, subscription))

	t.Log("Checking authorization directives...")
	ops := &auth.OperationsInput{
		auth.Get:    {fixPodQuery()},
		auth.List:   {fixPodsQuery()},
		auth.Delete: {fixDeletePodMutation()},
		auth.Update: {fixUpdatePodMutation("{\"\":\"\"}")},
	}
	AuthSuite.Run(t, ops)
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

func podEvent(eventType string, pod pod) PodEvent {
	return PodEvent{
		Type: eventType,
		Pod:  pod,
	}
}

func readPodEvent(sub *graphql.Subscription) (PodEvent, error) {
	type Response struct {
		PodEvent PodEvent
	}
	var podEvent Response
	err := sub.Next(&podEvent, tester.DefaultSubscriptionTimeout)

	return podEvent.PodEvent, err
}

func checkPodEvent(expected PodEvent, sub *graphql.Subscription) error {
	for {
		event, err := readPodEvent(sub)
		if err != nil {
			return err
		}
		if expected.Type == event.Type && expected.Pod.Name == event.Pod.Name {
			return nil
		}
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
	req.SetVar("namespace", testNamespace)

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
	req.SetVar("namespace", testNamespace)

	return req
}

func fixPodSubscription() *graphql.Request {
	query := `subscription ($namespace: String!) {
				podEvent(namespace: $namespace) {
					type
					pod {
						name
					}
				}
			}`
	req := graphql.NewRequest(query)
	req.SetVar("namespace", testNamespace)

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
	req.SetVar("namespace", testNamespace)
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
	req.SetVar("namespace", testNamespace)

	return req
}
