package kubernetes

import (
	"context"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func TestDeploymentProber_IsReady(t *testing.T) {
	tests := []struct {
		summary            string
		updatedReplicas    int32
		desiredScheduled   int32
		numberReady        int32
		observedGeneration int64
		desiredGeneration  int64
		expected           bool
	}{
		{summary: "all scheduled all ready", desiredScheduled: 1, numberReady: 1, updatedReplicas: 1, expected: true},
		{summary: "all scheduled one ready", desiredScheduled: 2, numberReady: 1, updatedReplicas: 2, expected: false},
		{summary: "all scheduled zero ready", desiredScheduled: 1, numberReady: 0, updatedReplicas: 1, expected: false},
		{summary: "scheduled mismatch", desiredScheduled: 1, numberReady: 2, updatedReplicas: 2, expected: false},
		{summary: "desired scheduled mismatch", desiredScheduled: 2, numberReady: 2, updatedReplicas: 1, expected: false},
		{summary: "generation mismatch", observedGeneration: 1, desiredGeneration: 2, expected: false},
	}

	for _, test := range tests {
		tc := test
		t.Run(tc.summary, func(t *testing.T) {
			t.Parallel()

			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "kyma-system", Generation: tc.desiredGeneration},
				Spec:       appsv1.DeploymentSpec{Replicas: &tc.desiredScheduled},
				Status: appsv1.DeploymentStatus{
					ReadyReplicas:      tc.numberReady,
					UpdatedReplicas:    tc.updatedReplicas,
					ObservedGeneration: tc.observedGeneration,
				},
			}
			fakeClient := fake.NewClientBuilder().WithObjects(deployment).Build()

			sut := DeploymentProber{fakeClient}
			ready, err := sut.IsReady(context.Background(), types.NamespacedName{Name: "foo", Namespace: "kyma-system"})

			require.NoError(t, err)
			require.Equal(t, tc.expected, ready)

		})
	}
}