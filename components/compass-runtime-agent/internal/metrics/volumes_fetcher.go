package metrics

import (
	"context"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
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

func (v *volumesFetcher) FetchPersistentVolumesCapacity() ([]PersistentVolumes, error) {
	persistentVolumes, err := v.persistentVolumesInterface.List(context.Background(), meta.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list persistent volumes")
	}

	persistentVolumesCapacity := make([]PersistentVolumes, 0)

	for _, persistentVolume := range persistentVolumes.Items {
		persistentVolumesCapacity = append(persistentVolumesCapacity, PersistentVolumes{
			Name:     persistentVolume.Name,
			Capacity: getCapacity(persistentVolume),
			Claim:    getClaim(persistentVolume),
		})
	}

	return persistentVolumesCapacity, nil
}

func getCapacity(pv v1.PersistentVolume) string {
	storage := pv.Spec.Capacity[v1.ResourceStorage]
	return (&storage).String()
}

func getClaim(pv v1.PersistentVolume) *Claim {
	if pv.Spec.ClaimRef == nil {
		return nil
	}
	return &Claim{
		Name:      pv.Spec.ClaimRef.Name,
		Namespace: pv.Spec.ClaimRef.Namespace,
	}
}
