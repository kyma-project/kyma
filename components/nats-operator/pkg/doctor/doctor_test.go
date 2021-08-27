package doctor

import (
	"context"
	"syscall"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	. "k8s.io/client-go/kubernetes/fake"

	"github.com/kyma-project/kyma/components/nats-operator/logger"
	. "github.com/kyma-project/kyma/components/nats-operator/pkg/client/natscluster/fake"
	"github.com/kyma-project/kyma/components/nats-operator/pkg/errors"
	. "github.com/kyma-project/kyma/components/nats-operator/pkg/testing"
	"github.com/nats-io/nats-operator/pkg/apis/nats/v1alpha2"
	"github.com/stretchr/testify/assert"
)

func TestDoctor(t *testing.T) {
	testCases := []struct {
		name                           string
		natsOperatorDeployment         *appsv1.Deployment
		natsOperatorPod                *v1.Pod
		natsServerPods                 []*v1.Pod
		natsCluster                    *v1alpha2.NatsCluster
		interval                       time.Duration
		interrupt                      time.Duration
		wantError                      bool
		wantRecoverable                bool
		wantOperatorPodDeleted         bool
		wantOperatorDeploymentScaledUp bool
	}{
		{
			name:                           "should not error if no resources found",
			natsOperatorDeployment:         nil,
			natsOperatorPod:                nil,
			natsServerPods:                 nil,
			natsCluster:                    nil,
			interval:                       time.Millisecond * 1,
			interrupt:                      time.Millisecond * 10,
			wantError:                      false,
			wantRecoverable:                false,
			wantOperatorPodDeleted:         false,
			wantOperatorDeploymentScaledUp: false,
		},
		{
			name:                           "should scale-up nats-operator deployment if spec replica is zero",
			natsOperatorDeployment:         GetDeployment(natsOperatorDeploymentName, namespace, DeploymentWithReplicas(0)),
			natsOperatorPod:                nil,
			natsServerPods:                 nil,
			natsCluster:                    nil,
			interval:                       time.Millisecond * 1,
			interrupt:                      time.Millisecond * 10,
			wantError:                      false,
			wantRecoverable:                false,
			wantOperatorPodDeleted:         false,
			wantOperatorDeploymentScaledUp: true,
		},
		{
			name:                           "should delete nats-operator pod if it is found but not running",
			natsOperatorDeployment:         GetDeployment(natsOperatorDeploymentName, namespace, DeploymentWithReplicas(1)),
			natsOperatorPod:                GetPod("nats-operator", namespace, PodWithLabel(labelKeyNatsOperator, labelValNatsOperator)),
			natsServerPods:                 nil,
			natsCluster:                    nil,
			interval:                       time.Millisecond * 1,
			interrupt:                      time.Millisecond * 10,
			wantError:                      false,
			wantRecoverable:                false,
			wantOperatorPodDeleted:         true,
			wantOperatorDeploymentScaledUp: false,
		},
		{
			name:                   "should delete nats-operator pod if actual running nats-servers is less than the desired",
			natsOperatorDeployment: GetDeployment(natsOperatorDeploymentName, namespace, DeploymentWithReplicas(1)),
			natsOperatorPod:        GetPod("nats-operator", namespace, PodWithLabel(labelKeyNatsOperator, labelValNatsOperator), PodWithPhase(v1.PodRunning)),
			natsServerPods: []*v1.Pod{
				GetPod("nats-server-0", namespace, PodWithLabel(labelKeyNatsCluster, labelValNatsCluster), PodWithPhase(v1.PodFailed)),
				GetPod("nats-server-1", namespace, PodWithLabel(labelKeyNatsCluster, labelValNatsCluster), PodWithPhase(v1.PodPending)),
				GetPod("nats-server-2", namespace, PodWithLabel(labelKeyNatsCluster, labelValNatsCluster), PodWithPhase(v1.PodRunning)),
				GetPod("nats-server-3", namespace, PodWithLabel(labelKeyNatsCluster, labelValNatsCluster), PodWithPhase(v1.PodRunning)),
			},
			natsCluster:                    GetNatsCluster(natsClusterName, namespace, 4),
			interval:                       time.Millisecond * 1,
			interrupt:                      time.Millisecond * 10,
			wantError:                      false,
			wantRecoverable:                false,
			wantOperatorPodDeleted:         true,
			wantOperatorDeploymentScaledUp: false,
		},
		{
			name:                   "should not delete nats-operator pod if actual running nats-servers is equal to the desired",
			natsOperatorDeployment: GetDeployment(natsOperatorDeploymentName, namespace, DeploymentWithReplicas(1)),
			natsOperatorPod:        GetPod("nats-operator", namespace, PodWithLabel(labelKeyNatsOperator, labelValNatsOperator), PodWithPhase(v1.PodRunning)),
			natsServerPods: []*v1.Pod{
				GetPod("nats-server-0", namespace, PodWithLabel(labelKeyNatsCluster, labelValNatsCluster), PodWithPhase(v1.PodRunning)),
				GetPod("nats-server-1", namespace, PodWithLabel(labelKeyNatsCluster, labelValNatsCluster), PodWithPhase(v1.PodRunning)),
				GetPod("nats-server-2", namespace, PodWithLabel(labelKeyNatsCluster, labelValNatsCluster), PodWithPhase(v1.PodRunning)),
				GetPod("nats-server-3", namespace, PodWithLabel(labelKeyNatsCluster, labelValNatsCluster), PodWithPhase(v1.PodRunning)),
				GetPod("nats-server-4", namespace, PodWithLabel(labelKeyNatsCluster, labelValNatsCluster), PodWithPhase(v1.PodRunning)),
			},
			natsCluster:                    GetNatsCluster(natsClusterName, namespace, 5),
			interval:                       time.Millisecond * 1,
			interrupt:                      time.Millisecond * 10,
			wantError:                      false,
			wantRecoverable:                false,
			wantOperatorPodDeleted:         false,
			wantOperatorDeploymentScaledUp: false,
		},
		{
			name:                   "should not delete nats-operator pod if actual running nats-servers is more than the desired",
			natsOperatorDeployment: GetDeployment(natsOperatorDeploymentName, namespace, DeploymentWithReplicas(1)),
			natsOperatorPod:        GetPod("nats-operator", namespace, PodWithLabel(labelKeyNatsOperator, labelValNatsOperator), PodWithPhase(v1.PodRunning)),
			natsServerPods: []*v1.Pod{
				GetPod("nats-server-0", namespace, PodWithLabel(labelKeyNatsCluster, labelValNatsCluster), PodWithPhase(v1.PodRunning)),
				GetPod("nats-server-1", namespace, PodWithLabel(labelKeyNatsCluster, labelValNatsCluster), PodWithPhase(v1.PodRunning)),
				GetPod("nats-server-2", namespace, PodWithLabel(labelKeyNatsCluster, labelValNatsCluster), PodWithPhase(v1.PodRunning)),
			},
			natsCluster:                    GetNatsCluster(natsClusterName, namespace, 1),
			interval:                       time.Millisecond * 1,
			interrupt:                      time.Millisecond * 10,
			wantError:                      false,
			wantRecoverable:                false,
			wantOperatorPodDeleted:         false,
			wantOperatorDeploymentScaledUp: false,
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
			doctor := New(k8sClient, natsClient, tc.interval, logger.New())

			// when
			stopDoctorAfter(doctor, tc.interrupt)
			err := doctor.Start(context.Background())
			gotError := err != nil

			// expect
			assert.NotNil(t, doctor)
			assert.Equal(t, tc.wantError, gotError)
			if tc.wantError {
				gotRecoverable := errors.IsRecoverable(err)
				assert.Equal(t, tc.wantRecoverable, gotRecoverable)
			}
			if tc.wantOperatorPodDeleted {
				assert.Nil(t, doctor.state.natsOperatorPod)
			} else {
				assert.Equal(t, tc.natsOperatorPod, doctor.state.natsOperatorPod)
			}
			if tc.wantOperatorDeploymentScaledUp {
				assert.Equal(t, int32(1), *doctor.state.natsOperatorDeployment.Spec.Replicas)
			}
		})
	}
}

// stopDoctorAfter sends an interrupt signal to the doctor instance after the given duration.
func stopDoctorAfter(doctor *Doctor, duration time.Duration) {
	go func() {
		time.Sleep(duration)
		doctor.stop <- syscall.SIGINT
	}()
}
