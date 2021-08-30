package doctor

import (
	"context"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	. "k8s.io/client-go/kubernetes/fake"

	. "github.com/kyma-project/kyma/components/nats-operator/pkg/client/natscluster/fake"
	"github.com/kyma-project/kyma/components/nats-operator/pkg/errors"
	. "github.com/kyma-project/kyma/components/nats-operator/pkg/testing"
	"github.com/nats-io/nats-operator/pkg/apis/nats/v1alpha2"
	"github.com/stretchr/testify/assert"
)

func TestIsNatsBackend(t *testing.T) {
	testCases := []struct {
		name        string
		secrets     []*v1.Secret
		natsBackend bool
	}{
		{
			name: "one beb secret",
			secrets: []*v1.Secret{
				GetSecret("secret", namespace, SecretWithLabel(labelKeyEventingBackend, labelValEventingBackendBeb)),
			},
			natsBackend: false,
		},
		{
			name: "multiple beb secrets",
			secrets: []*v1.Secret{
				GetSecret("secret-0", namespace, SecretWithLabel(labelKeyEventingBackend, labelValEventingBackendBeb)),
				GetSecret("secret-1", namespace, SecretWithLabel(labelKeyEventingBackend, labelValEventingBackendBeb)),
				GetSecret("secret-2", namespace, SecretWithLabel(labelKeyEventingBackend, labelValEventingBackendBeb)),
			},
			natsBackend: false,
		},
		{
			name:        "one non-beb secret",
			secrets:     []*v1.Secret{GetSecret("secret", namespace)},
			natsBackend: true,
		},
		{
			name:        "no secrets",
			secrets:     nil,
			natsBackend: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// resources
			secrets := SecretsToRuntimeObjects(tc.secrets)

			// clients
			k8sClient := NewSimpleClientset(secrets...)

			// given
			state := newState(k8sClient, nil)

			// when
			natsBackend, err := state.isNatsBackend(context.Background())

			// expect
			assert.Nil(t, err)
			assert.Equal(t, tc.natsBackend, natsBackend)
		})
	}
}

func TestComputeState(t *testing.T) {
	testCases := []struct {
		name                   string
		natsOperatorDeployment *appsv1.Deployment
		natsOperatorPod        *v1.Pod
		natsServerPods         []*v1.Pod
		natsCluster            *v1alpha2.NatsCluster
		wantError              bool
		wantRecoverable        bool
		wantNatsServersActual  int
		wantNatsServersDesired int
	}{
		{
			name:                   "nats-cluster custom-resource not found",
			natsCluster:            nil,
			natsOperatorDeployment: nil,
			natsOperatorPod:        nil,
			natsServerPods:         nil,
			wantError:              true,
			wantRecoverable:        true,
			wantNatsServersActual:  0,
			wantNatsServersDesired: 0,
		},
		{
			name:                   "nats-operator deployment not found",
			natsCluster:            GetNatsCluster(natsClusterName, namespace, 0),
			natsOperatorDeployment: nil,
			natsOperatorPod:        nil,
			natsServerPods:         nil,
			wantError:              true,
			wantRecoverable:        true,
			wantNatsServersActual:  0,
			wantNatsServersDesired: 0,
		},
		{
			name:                   "nats-operator pod not found",
			natsCluster:            GetNatsCluster(natsClusterName, namespace, 0),
			natsOperatorDeployment: GetDeployment(natsOperatorDeploymentName, namespace),
			natsOperatorPod:        nil,
			natsServerPods:         nil,
			wantError:              true,
			wantRecoverable:        true,
			wantNatsServersActual:  0,
			wantNatsServersDesired: 0,
		},
		{
			name:                   "nats-server pods not found",
			natsCluster:            GetNatsCluster(natsClusterName, namespace, 0),
			natsOperatorDeployment: GetDeployment(natsOperatorDeploymentName, namespace),
			natsOperatorPod:        GetPod("nats-operator", namespace, PodWithLabel(labelKeyNatsOperator, labelValNatsOperator)),
			natsServerPods:         nil,
			wantError:              true,
			wantRecoverable:        true,
			wantNatsServersActual:  0,
			wantNatsServersDesired: 0,
		},
		{
			name:                   "all resources found",
			natsCluster:            GetNatsCluster(natsClusterName, namespace, 5),
			natsOperatorDeployment: GetDeployment(natsOperatorDeploymentName, namespace, DeploymentWithReplicas(1)),
			natsOperatorPod:        GetPod("nats-operator", namespace, PodWithLabel(labelKeyNatsOperator, labelValNatsOperator), PodWithPhase(v1.PodRunning)),
			natsServerPods: []*v1.Pod{
				GetPod("nats-server-0", namespace, PodWithLabel(labelKeyNatsCluster, labelValNatsCluster)),
				GetPod("nats-server-1", namespace, PodWithLabel(labelKeyNatsCluster, labelValNatsCluster), PodWithPhase(v1.PodPending)),
				GetPod("nats-server-2", namespace, PodWithLabel(labelKeyNatsCluster, labelValNatsCluster), PodWithPhase(v1.PodRunning)),
				GetPod("nats-server-3", namespace, PodWithLabel(labelKeyNatsCluster, labelValNatsCluster), PodWithPhase(v1.PodRunning)),
				GetPod("nats-server-4", namespace, PodWithLabel(labelKeyNatsCluster, labelValNatsCluster), PodWithPhase(v1.PodRunning)),
			},
			wantError:              false,
			wantRecoverable:        false,
			wantNatsServersActual:  3,
			wantNatsServersDesired: 5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// resources
			pods := PodsToRuntimeObjects(tc.natsServerPods)
			pods = append(pods, GetPodIfNil(tc.natsOperatorPod), GetDeploymentIfNil(tc.natsOperatorDeployment))
			natsClusterList := GetNatsClusterList(*GetNatsClusterIfNil(tc.natsCluster))

			// clients
			k8sClient := NewSimpleClientset(pods...)
			natsClient := NewClientOrDie(natsClusterList)

			// given
			state := newState(k8sClient, natsClient)

			// when
			err := state.compute(context.Background())
			gotError := err != nil

			// expect
			assert.NotNil(t, state)
			if assert.Equal(t, tc.wantError, gotError) && tc.wantError {
				gotRecoverable := errors.IsRecoverable(err)
				assert.Equal(t, tc.wantRecoverable, gotRecoverable)
			}
			assert.Equal(t, tc.wantNatsServersDesired, state.natsServersDesired)
			assert.Equal(t, tc.natsOperatorDeployment, state.natsOperatorDeployment)
			assert.Equal(t, tc.natsOperatorPod, state.natsOperatorPod)
			assert.Equal(t, tc.wantNatsServersActual, state.natsServersActual)
		})
	}
}

func TestComputeNatsServersDesired(t *testing.T) {
	testCases := []struct {
		name                   string
		natsCluster            *v1alpha2.NatsCluster
		wantError              bool
		wantRecoverable        bool
		wantNatsServersDesired int
	}{
		{
			name:                   "nats-cluster custom-resource not found",
			natsCluster:            nil,
			wantError:              true,
			wantRecoverable:        true,
			wantNatsServersDesired: 0,
		},
		{
			name:                   "nats-cluster custom-resource found",
			natsCluster:            GetNatsCluster(natsClusterName, namespace, 1),
			wantError:              false,
			wantNatsServersDesired: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// resources
			natsClusterList := GetNatsClusterList(*GetNatsClusterIfNil(tc.natsCluster))

			// clients
			natsClient := NewClientOrDie(natsClusterList)

			// given
			state := newState(nil, natsClient)

			// when
			err := state.computeNatsServersDesired(context.Background())
			gotError := err != nil

			// expect
			assert.NotNil(t, state)
			assert.Equal(t, tc.wantNatsServersDesired, state.natsServersDesired)
			assert.Equal(t, tc.wantError, gotError)
			if tc.wantError {
				gotRecoverable := errors.IsRecoverable(err)
				assert.Equal(t, tc.wantRecoverable, gotRecoverable)
			}
		})
	}
}

func TestComputeNatsOperatorDeployment(t *testing.T) {
	testCases := []struct {
		name            string
		deployment      *appsv1.Deployment
		wantError       bool
		wantRecoverable bool
	}{
		{
			name:            "nats-operator deployment not found",
			deployment:      nil,
			wantError:       true,
			wantRecoverable: true,
		},
		{
			name:       "nats-operator deployment found",
			deployment: GetDeployment(natsOperatorDeploymentName, namespace),
			wantError:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// resources
			deployment := GetDeploymentIfNil(tc.deployment)

			// clients
			k8sClient := NewSimpleClientset(deployment)

			// given
			state := newState(k8sClient, nil)

			// when
			err := state.computeNatsOperatorDeployment(context.Background())
			gotError := err != nil

			// expect
			assert.NotNil(t, state)
			assert.Equal(t, tc.wantError, gotError)
			assert.Equal(t, tc.deployment, state.natsOperatorDeployment)
			if tc.wantError {
				gotRecoverable := errors.IsRecoverable(err)
				assert.Equal(t, tc.wantRecoverable, gotRecoverable)
			}
		})
	}
}

func TestComputeNatsOperatorPod(t *testing.T) {
	testCases := []struct {
		name            string
		pod             *v1.Pod
		wantError       bool
		wantRecoverable bool
	}{
		{
			name:            "nats-operator pod not found",
			pod:             nil,
			wantError:       true,
			wantRecoverable: true,
		},
		{
			name:            "nats-operator pod found but not running",
			pod:             GetPod("nats-operator", namespace, PodWithLabel(labelKeyNatsOperator, labelValNatsOperator)),
			wantError:       true,
			wantRecoverable: true,
		},
		{
			name:      "nats-operator pod found and running",
			pod:       GetPod("nats-operator", namespace, PodWithLabel(labelKeyNatsOperator, labelValNatsOperator), PodWithPhase(v1.PodRunning)),
			wantError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// resources
			pod := GetPodIfNil(tc.pod)

			// clients
			k8sClient := NewSimpleClientset(pod)

			// given
			state := newState(k8sClient, nil)

			// when
			err := state.computeNatsOperatorPod(context.Background())
			gotError := err != nil

			// expect
			assert.NotNil(t, state)
			assert.Equal(t, tc.wantError, gotError)
			assert.Equal(t, tc.pod, state.natsOperatorPod)
			if tc.wantError {
				gotRecoverable := errors.IsRecoverable(err)
				assert.Equal(t, tc.wantRecoverable, gotRecoverable)
			}
		})
	}
}

func TestComputeNatsServersActual(t *testing.T) {
	testCases := []struct {
		name                  string
		pods                  []*v1.Pod
		wantError             bool
		wantRecoverable       bool
		wantNatsServersActual int
	}{
		{
			name:                  "nats-server pods not found",
			pods:                  nil,
			wantError:             true,
			wantRecoverable:       true,
			wantNatsServersActual: 0,
		},
		{
			name: "nats-server pods found with different phases",
			pods: []*v1.Pod{
				GetPod("nats-server-0", namespace, PodWithLabel(labelKeyNatsCluster, labelValNatsCluster), PodWithPhase(v1.PodFailed)),
				GetPod("nats-server-1", namespace, PodWithLabel(labelKeyNatsCluster, labelValNatsCluster), PodWithPhase(v1.PodRunning)),
			},
			wantError:             false,
			wantNatsServersActual: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// resources
			pods := PodsToRuntimeObjects(tc.pods)

			// clients
			k8sClient := NewSimpleClientset(pods...)

			// given
			state := newState(k8sClient, nil)

			// when
			err := state.computeNatsServersActual(context.Background())
			gotError := err != nil

			// expect
			assert.NotNil(t, state)
			assert.Equal(t, tc.wantNatsServersActual, state.natsServersActual)
			assert.Equal(t, tc.wantError, gotError)
			if tc.wantError {
				gotRecoverable := errors.IsRecoverable(err)
				assert.Equal(t, tc.wantRecoverable, gotRecoverable)
			}
		})
	}
}

func TestIsPodListEmpty(t *testing.T) {
	testCases := []struct {
		name      string
		podList   *v1.PodList
		wantEmpty bool
	}{
		{
			name:      "nil pod list",
			podList:   nil,
			wantEmpty: true,
		},
		{
			name:      "empty pod list",
			podList:   &v1.PodList{},
			wantEmpty: true,
		},
		{
			name: "non-empty pod list",
			podList: &v1.PodList{
				Items: []v1.Pod{
					*GetPod("pod-0", namespace),
				},
			},
			wantEmpty: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// when
			gotEmpty := isPodListEmpty(tc.podList)

			// expect
			assert.Equal(t, tc.wantEmpty, gotEmpty)
		})
	}
}
