package metrics

import (
	"github.com/pkg/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	core "k8s.io/client-go/kubernetes/typed/core/v1"
)

type VolumesFetcher interface {
	FetchPersistentVolumesCapacity() ([]PersistentVolumes, error)
}

type volumesFetcher struct {
	persistentVolumesInterface core.PersistentVolumeInterface
}

func newVolumesFetcher(clientset kubernetes.Interface) VolumesFetcher {
	return &volumesFetcher{
		persistentVolumesInterface: clientset.CoreV1().PersistentVolumes(),
	}
}

func (r *volumesFetcher) FetchPersistentVolumesCapacity() ([]PersistentVolumes, error) {
	persistentVolumes, err := r.persistentVolumesInterface.List(meta.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list persistent volumes")
	}

	persistentVolumesCapacity := make([]PersistentVolumes, 0)

	for _, persistentVolume := range persistentVolumes.Items {
		persistentVolumesCapacity = append(persistentVolumesCapacity, PersistentVolumes{
			Name:      persistentVolume.Name,
			Namespace: persistentVolume.Namespace,
			Capacity: ResourceGroup{
				CPU:              persistentVolume.Spec.Capacity.Cpu().String(),
				EphemeralStorage: persistentVolume.Spec.Capacity.StorageEphemeral().String(),
				Memory:           persistentVolume.Spec.Capacity.Memory().String(),
				Pods:             persistentVolume.Spec.Capacity.Pods().String(),
			},
		})
	}

	return persistentVolumesCapacity, nil
}
