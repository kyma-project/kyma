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
				Name: "somename",
			},
			Spec: corev1.PersistentVolumeSpec{
				Capacity: corev1.ResourceList{
					corev1.ResourceStorage: *resource.NewQuantity(1, resource.BinarySI),
				},
				ClaimRef: &corev1.ObjectReference{
					Namespace: "claimnamespace",
					Name:      "claimname",
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
		assert.Equal(t, "1", volumes[0].Capacity)
		require.NotNil(t, volumes[0].Claim)
		assert.Equal(t, "claimnamespace", volumes[0].Claim.Namespace)
		assert.Equal(t, "claimname", volumes[0].Claim.Name)
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

	t.Run("should return 0 capacity if none is allocated", func(t *testing.T) {
		// given
		resourcesClientset := kubernetesFake.NewSimpleClientset(&corev1.PersistentVolume{
			Spec: corev1.PersistentVolumeSpec{
				Capacity: corev1.ResourceList{},
			},
		})
		volumesFetcher := newVolumesFetcher(resourcesClientset)

		// when
		volumes, err := volumesFetcher.FetchPersistentVolumesCapacity()
		require.NoError(t, err)

		// then
		require.Equal(t, 1, len(volumes))
		assert.Equal(t, "0", volumes[0].Capacity)
	})

	t.Run("should return nil claim if none is bound", func(t *testing.T) {
		// given
		resourcesClientset := kubernetesFake.NewSimpleClientset(&corev1.PersistentVolume{})
		volumesFetcher := newVolumesFetcher(resourcesClientset)

		// when
		volumes, err := volumesFetcher.FetchPersistentVolumesCapacity()
		require.NoError(t, err)

		// then
		require.Equal(t, 1, len(volumes))
		assert.Nil(t, volumes[0].Claim)
	})
}
