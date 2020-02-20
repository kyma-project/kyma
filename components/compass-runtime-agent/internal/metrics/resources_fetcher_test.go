package metrics

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubernetesFake "k8s.io/client-go/kubernetes/fake"
	"testing"
)

func Test_FetchNodesResources(t *testing.T) {
	t.Run("should fetch nodes resources", func(t *testing.T) {
		// given
		resourcesClientset := kubernetesFake.NewSimpleClientset(&corev1.Node{
			ObjectMeta: v1.ObjectMeta{
				Name:   "somename",
				Labels: map[string]string{"beta.kubernetes.io/instance-type": "somelabel"},
			},
			Status: corev1.NodeStatus{
				Capacity: corev1.ResourceList{
					corev1.ResourceCPU:              *resource.NewQuantity(1, resource.DecimalSI),
					corev1.ResourceMemory:           *resource.NewQuantity(1, resource.BinarySI),
					corev1.ResourceEphemeralStorage: *resource.NewQuantity(1, resource.BinarySI),
					corev1.ResourcePods:             *resource.NewQuantity(1, resource.DecimalSI),
				},
			},
		})
		resourcesFetcher := newResourcesFetcher(resourcesClientset)

		// when
		resources, err := resourcesFetcher.FetchNodesResources()
		require.NoError(t, err)

		// then
		require.Equal(t, 1, len(resources))
		assert.Equal(t, "somename", resources[0].Name)
		assert.Equal(t, "somelabel", resources[0].InstanceType)
		assert.Equal(t, "1", resources[0].Capacity.CPU)
		assert.Equal(t, "1", resources[0].Capacity.Memory)
		assert.Equal(t, "1", resources[0].Capacity.EphemeralStorage)
		assert.Equal(t, "1", resources[0].Capacity.Pods)
	})

	t.Run("should not fail if no nodes", func(t *testing.T) {
		// given
		resourcesClientset := kubernetesFake.NewSimpleClientset()
		resourcesFetcher := newResourcesFetcher(resourcesClientset)

		// when
		resources, err := resourcesFetcher.FetchNodesResources()
		require.NoError(t, err)

		// then
		assert.Equal(t, 0, len(resources))
	})
}
