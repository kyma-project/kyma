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
		summary          string
		desiredScheduled int32
		numberReady      int32
		expected         bool
	}{
		{summary: "all scheduled all ready", desiredScheduled: 1, numberReady: 1, expected: true},
		{summary: "all scheduled one ready", desiredScheduled: 2, numberReady: 1, expected: false},
		{summary: "all scheduled zero ready", desiredScheduled: 1, numberReady: 0, expected: false},
	}

	for _, test := range tests {
		tc := test
		t.Run(tc.summary, func(t *testing.T) {

			t.Parallel()

			matchLabels := make(map[string]string)
			matchLabels["test.deployment.name"] = "test-deployment"

			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "kyma-system"},
				Spec: appsv1.DeploymentSpec{
					Replicas: &tc.desiredScheduled,
					Selector: &metav1.LabelSelector{MatchLabels: matchLabels},
				},
				Status: appsv1.DeploymentStatus{
					ReadyReplicas: tc.numberReady,
				},
			}

			rs := &appsv1.ReplicaSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "foo",
					Namespace:       "kyma-system",
					Labels:          deployment.Spec.Selector.MatchLabels,
					OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(deployment, deployment.GroupVersionKind())},
				},
				Spec: appsv1.ReplicaSetSpec{
					Selector: deployment.Spec.Selector,
					Replicas: &tc.desiredScheduled,
					Template: deployment.Spec.Template,
				},
				Status: appsv1.ReplicaSetStatus{
					ReadyReplicas: tc.numberReady,
					Replicas:      tc.numberReady,
				},
			}

			itemList := make([]appsv1.ReplicaSet, 1)

			itemList = append(itemList, *rs)
			rsList := &appsv1.ReplicaSetList{
				Items: itemList,
			}

			fakeClient := fake.NewClientBuilder().WithObjects(deployment).WithLists(rsList).Build()

			sut := DeploymentProber{fakeClient}
			ready, err := sut.IsReady(context.Background(), types.NamespacedName{Name: "foo", Namespace: "kyma-system"})

			require.NoError(t, err)
			require.Equal(t, tc.expected, ready)

		})
	}
}
