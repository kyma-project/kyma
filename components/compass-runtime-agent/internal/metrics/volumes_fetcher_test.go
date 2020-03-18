package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubernetesFake "k8s.io/client-go/kubernetes/fake"
)

func Test_FetchPersistentVolumesCapacity(t *testing.T) {
	t.Run("should fetch persistent volumes capacity", func(t *testing.T) {
		// given
		resourcesClientset := kubernetesFake.NewSimpleClientset(&corev1.PersistentVolume{
			ObjectMeta: v1.ObjectMeta{
				Name:      "somename",
				Namespace: "somenamespace",
			},
			Spec: corev1.PersistentVolumeSpec{
				Capacity: corev1.ResourceList{
					corev1.ResourceCPU:              *resource.NewQuantity(1, resource.DecimalSI),
					corev1.ResourceMemory:           *resource.NewQuantity(1, resource.BinarySI),
					corev1.ResourceEphemeralStorage: *resource.NewQuantity(1, resource.BinarySI),
					corev1.ResourcePods:             *resource.NewQuantity(1, resource.DecimalSI),
				},
			},
		})
		volumesFetcher := newVolumesFetcher(resourcesClientset)

		// when
		volumes, err := volumesFetcher.FetchPersistentVolumesCapacity()
		require.NoError(t, err)

		// then
		require.Equal(t, 1, len(volumes))
		assert.Equal(t, "somename", volumes[0].Name)
		assert.Equal(t, "somenamespace", volumes[0].Namespace)
		assert.Equal(t, "1", volumes[0].Capacity.CPU)
		assert.Equal(t, "1", volumes[0].Capacity.Memory)
		assert.Equal(t, "1", volumes[0].Capacity.EphemeralStorage)
		assert.Equal(t, "1", volumes[0].Capacity.Pods)
	})

	t.Run("should not fail if no persistent volumes", func(t *testing.T) {
		// given
		resourcesClientset := kubernetesFake.NewSimpleClientset()
		volumesFetcher := newVolumesFetcher(resourcesClientset)

		// when
		volumes, err := volumesFetcher.FetchPersistentVolumesCapacity()
		require.NoError(t, err)

		// then
		assert.Equal(t, 0, len(volumes))
	})
}
