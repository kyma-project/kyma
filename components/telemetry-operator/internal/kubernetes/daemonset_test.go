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

func TestDaemonSetProber(t *testing.T) {
	tests := []struct {
		summary            string
		updatedScheduled   int32
		desiredScheduled   int32
		numberReady        int32
		observedGeneration int64
		desiredGeneration  int64
		expected           bool
	}{
		{summary: "all scheduled all ready", desiredScheduled: 3, numberReady: 3, updatedScheduled: 3, expected: true},
		{summary: "all scheduled one ready", desiredScheduled: 3, numberReady: 1, updatedScheduled: 3, expected: false},
		{summary: "all scheduled zero ready", desiredScheduled: 3, numberReady: 0, updatedScheduled: 3, expected: false},
		{summary: "scheduled mismatch", desiredScheduled: 1, numberReady: 3, updatedScheduled: 3, expected: false},
		{summary: "desired scheduled mismatch", desiredScheduled: 3, numberReady: 3, updatedScheduled: 1, expected: false},
		{summary: "generation mismatch", observedGeneration: 1, desiredGeneration: 2, expected: false},
	}

	for _, test := range tests {
		tc := test
		t.Run(tc.summary, func(t *testing.T) {
			t.Parallel()

			daemonSet := &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "kyma-system", Generation: tc.desiredGeneration},
				Status: appsv1.DaemonSetStatus{
					DesiredNumberScheduled: tc.desiredScheduled,
					NumberReady:            tc.numberReady,
					UpdatedNumberScheduled: tc.updatedScheduled,
					ObservedGeneration:     tc.observedGeneration,
				},
			}

			fakeClient := fake.NewClientBuilder().WithObjects(daemonSet).Build()

			sut := DaemonSetProber{fakeClient}
			ready, err := sut.IsReady(context.Background(), types.NamespacedName{Name: "foo", Namespace: "kyma-system"})

			require.NoError(t, err)
			require.Equal(t, tc.expected, ready)
		})
	}
}

func TestSetAnnotation(t *testing.T) {
	daemonSet := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "kyma-system"},
	}

	fakeClient := fake.NewClientBuilder().WithObjects(daemonSet).Build()

	sut := DaemonSetAnnotator{fakeClient}

	err := sut.SetAnnotation(context.Background(), types.NamespacedName{Name: "foo", Namespace: "kyma-system"}, "foo", "bar")
	require.NoError(t, err)

	var updatedDaemonSet appsv1.DaemonSet
	_ = fakeClient.Get(context.Background(), types.NamespacedName{Name: "foo", Namespace: "kyma-system"}, &updatedDaemonSet)
	require.Len(t, updatedDaemonSet.Spec.Template.Annotations, 1)
	require.Contains(t, updatedDaemonSet.Spec.Template.Annotations, "foo")
	require.Equal(t, updatedDaemonSet.Spec.Template.Annotations["foo"], "bar")
}
